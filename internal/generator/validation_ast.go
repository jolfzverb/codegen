package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"

	"github.com/sintoniastrategy/validgo-gen/internal/generator/astbuilder"
)

func GetSchemaValidators(schema *openapi3.SchemaRef) []string {
	var validateTags []string
	switch {
	case schema.Value.Type.Permits(openapi3.TypeString):
		if schema.Value.MinLength > 0 {
			validateTags = append(validateTags, "min="+strconv.FormatUint(schema.Value.MinLength, 10))
		}
		if schema.Value.MaxLength != nil {
			validateTags = append(validateTags, "max="+strconv.FormatUint(*schema.Value.MaxLength, 10))
		}
		if schema.Value.Pattern != "" {
			slog.Warn("pattern validator is not supported", slog.String("pattern", schema.Value.Pattern))
		}
		if len(schema.Value.Enum) > 0 {
			enumStrValues := make([]string, 0, len(schema.Value.Enum))
			for _, enumValue := range schema.Value.Enum {
				var enumStrValue string
				if strValue, ok := enumValue.(string); ok {
					enumStrValue = strValue
				} else {
					slog.Warn("enum value is not a string", slog.Any("value", enumValue))
					enumStrValue = fmt.Sprintf("%v", enumValue)
				}
				if enumStrValue == "" || strings.Contains(enumStrValue, " ") {
					enumStrValue = "'" + enumStrValue + "'"
				}
				enumStrValues = append(enumStrValues, enumStrValue)
			}
			joinedEnum := strings.Join(enumStrValues, " ")
			validateTags = append(validateTags, "oneof="+joinedEnum)
		}
		switch schema.Value.Format {
		case "ip":
			validateTags = append(validateTags, "ip")
		case "ipv4":
			validateTags = append(validateTags, "ipv4")
		case "ipv6":
			validateTags = append(validateTags, "ipv6")
		case "email":
			validateTags = append(validateTags, "email")
		}

	case schema.Value.Type.Permits(openapi3.TypeInteger):
		if schema.Value.Min != nil {
			validateTags = append(validateTags, "min="+fmt.Sprint(*schema.Value.Min))
		}
		if schema.Value.Max != nil {
			validateTags = append(validateTags, "max="+fmt.Sprint(*schema.Value.Max))
		}
		if schema.Value.MultipleOf != nil {
			slog.Warn("multipleOf validator is not supported")
		}
		if schema.Value.ExclusiveMax {
			slog.Warn("exclusiveMax validator is not supported")
		}
		if schema.Value.ExclusiveMin {
			slog.Warn("exclusiveMin validator is not supported")
		}
		if len(schema.Value.Enum) > 0 {
			enumStrValues := make([]string, 0, len(schema.Value.Enum))
			for _, enumValue := range schema.Value.Enum {
				enumStrValue := fmt.Sprintf("%v", enumValue)
				enumStrValues = append(enumStrValues, enumStrValue)
			}
			joinedEnum := strings.Join(enumStrValues, " ")
			validateTags = append(validateTags, "oneof="+joinedEnum)
		}

	case schema.Value.Type.Permits(openapi3.TypeNumber):
		if schema.Value.Min != nil {
			validateTags = append(validateTags, "min="+fmt.Sprint(*schema.Value.Min))
		}
		if schema.Value.Max != nil {
			validateTags = append(validateTags, "max="+fmt.Sprint(*schema.Value.Max))
		}
		if schema.Value.MultipleOf != nil {
			slog.Warn("multipleOf validator is not supported")
		}
		if schema.Value.ExclusiveMax {
			slog.Warn("exclusiveMax validator is not supported")
		}
		if schema.Value.ExclusiveMin {
			slog.Warn("exclusiveMin validator is not supported")
		}
		if len(schema.Value.Enum) > 0 {
			enumStrValues := make([]string, 0, len(schema.Value.Enum))
			for _, enumValue := range schema.Value.Enum {
				enumStrValue := fmt.Sprintf("%v", enumValue)
				enumStrValues = append(enumStrValues, enumStrValue)
			}
			joinedEnum := strings.Join(enumStrValues, " ")
			validateTags = append(validateTags, "oneof="+joinedEnum)
		}

	case schema.Value.Type.Permits(openapi3.TypeArray):
		if schema.Value.MinItems > 0 {
			validateTags = append(validateTags, "min="+strconv.FormatUint(schema.Value.MinItems, 10))
		}
		if schema.Value.MaxItems != nil {
			validateTags = append(validateTags, "max="+strconv.FormatUint(*schema.Value.MaxItems, 10))
		}
		if schema.Value.UniqueItems {
			validateTags = append(validateTags, "unique")
		}
		validateTags = append(validateTags, "dive")
		itemsValidators := GetSchemaValidators(schema.Value.Items)
		validateTags = append(validateTags, itemsValidators...)
	}

	return validateTags
}

func (g *Generator) GetValidateFuncStmt(typeName string, ref string) ast.Expr {
	validateFuncName := "Validate" + typeName + "JSON"

	if ref == "" || !refIsExternal(ref) {
		return I(validateFuncName)
	}

	filename := parseFilenameFromRef(ref)
	if filename == "" {
		return I(validateFuncName)
	}

	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return I(validateFuncName)
	}

	validateFuncName = "Validate" + parts[len(parts)-1] + "JSON"

	g.YAMLFilesToProcess = append(g.YAMLFilesToProcess, g.GetYAMLFilePath(filename))
	g.AddHandlersImport(g.GetHandlersImportForFile(filename))
	modelName := g.GetModelName(filename)
	return Sel(I(modelName), validateFuncName)
}

func (g *Generator) AddContainsNullIfNeeded() {
	if g.HandlersFile.hasContainsNullMethod {
		return
	}

	g.HandlersFile.hasContainsNullMethod = true

	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("temp")},
						Type:  I("any"),
					},
				},
			},
		}).
		AddStmt(astbuilder.DefineCall("err", Sel(I("json"), "Unmarshal"), I("data"), Amp(I("temp")))).
		AddStmt(astbuilder.If(Ne(I("err"), I("nil"))).WithBody(astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.Return1(I("false"))))).
		AddStmt(astbuilder.Return1(&ast.BinaryExpr{X: I("temp"), Op: token.EQL, Y: I("nil")}))

	fn := astbuilder.Function("containsNull").
		AddParam(astbuilder.SelectorField("data", "json", "RawMessage")).
		AddResult(astbuilder.BoolField("")).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)
	g.AddHandlersImport("encoding/json")
}

func (g *Generator) AddObjectValidate(modelName string, schema *openapi3.SchemaRef) error {
	const op = "generator.AddObjectValidate"
	requiredFieldsMap := make(map[string]bool, 0)
	nullableFields := make([]string, 0)
	objectFields := make(map[string]ast.Expr, 0)

	for _, requiredField := range schema.Value.Required {
		requiredFieldsMap[requiredField] = true
	}

	for fieldName, fieldSchema := range schema.Value.Properties {
		if fieldSchema.Value == nil {
			continue
		}
		if fieldSchema.Value.Nullable && requiredFieldsMap[fieldName] {
			nullableFields = append(nullableFields, fieldName)
		}
		if fieldSchema.Value.Type.Permits(openapi3.TypeObject) {
			fieldType, err := g.GetFieldTypeFromSchema(modelName, fieldName, fieldSchema)
			if err != nil {
				return errors.Wrap(err, op)
			}
			objectFields[fieldName] = g.GetValidateFuncStmt(fieldType, fieldSchema.Ref)
		}
		if fieldSchema.Value.Type.Permits(openapi3.TypeArray) {
			if fieldSchema.Value.Items != nil {
				itemsType := g.getMostNestedArrayItemType(fieldSchema.Value.Items)
				if itemsType != nil && itemsType.Permits(openapi3.TypeObject) {
					fieldType, err := g.GetFieldTypeFromSchema(modelName, fieldName, fieldSchema)
					if err != nil {
						return errors.Wrap(err, op)
					}
					objectFields[fieldName] = g.GetValidateFuncStmt(fieldType, fieldSchema.Ref)
				}
			}
		}
	}
	requiredFields := make([]string, 0, len(requiredFieldsMap))
	for fieldName := range requiredFieldsMap {
		requiredFields = append(requiredFields, fieldName)
	}
	sort.Strings(requiredFields)
	sort.Strings(nullableFields)

	bodyBuilder := astbuilder.NewBodyBuilder()

	if len(requiredFields) > 0 {
		requiredFieldsElts := make([]ast.Expr, 0, len(requiredFields))
		for _, fieldName := range requiredFields {
			requiredFieldsElts = append(requiredFieldsElts, &ast.KeyValueExpr{Key: Str(fieldName), Value: I("true")})
		}
		bodyBuilder.AddStmt(astbuilder.Define(I("requiredFields"), &ast.CompositeLit{
			Type: &ast.MapType{Key: I("string"), Value: I("bool")},
			Elts: requiredFieldsElts,
		}))

		nullableFieldsElts := make([]ast.Expr, 0, len(nullableFields))
		for _, fieldName := range nullableFields {
			nullableFieldsElts = append(nullableFieldsElts, &ast.KeyValueExpr{Key: Str(fieldName), Value: I("true")})
		}
		bodyBuilder.AddStmt(astbuilder.Define(I("nullableFields"), &ast.CompositeLit{
			Type: &ast.MapType{Key: I("string"), Value: I("bool")},
			Elts: nullableFieldsElts,
		}))
	}

	if len(requiredFields) > 0 || len(objectFields) > 0 {
		bodyBuilder.AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("obj")},
						Type:  &ast.MapType{Key: I("string"), Value: Sel(I("json"), "RawMessage")},
					},
				},
			},
		})
		bodyBuilder.
			AddStmt(astbuilder.DefineCall("err", Sel(I("json"), "Unmarshal"), I("jsonData"), Amp(I("obj")))).
			AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().AddStmt(astbuilder.Return1(I("err")))))
	}

	if len(requiredFields) > 0 || len(objectFields) > 0 {
		bodyBuilder.AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok:   token.VAR,
				Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{I("val")}, Type: Sel(I("json"), "RawMessage")}},
			},
		})
		bodyBuilder.AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok:   token.VAR,
				Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{I("exists")}, Type: I("bool")}},
			},
		})
	}

	if len(requiredFields) > 0 {
		rangeBody := astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.NewAssignBuilder().Lhs(I("val"), I("exists")).Rhs(&ast.IndexExpr{X: I("obj"), Index: I("field")})).
			AddStmt(astbuilder.If(&ast.UnaryExpr{Op: token.NOT, X: I("exists")}).WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return1(&ast.CallExpr{
					Fun: Sel(I("errors"), "New"),
					Args: []ast.Expr{&ast.BinaryExpr{
						X:  &ast.BinaryExpr{X: Str("field "), Op: token.ADD, Y: I("field")},
						Op: token.ADD,
						Y:  Str(" is required"),
					}},
				})))).
			AddStmt(astbuilder.If(&ast.BinaryExpr{
				X:  &ast.UnaryExpr{Op: token.NOT, X: &ast.IndexExpr{X: I("nullableFields"), Index: I("field")}},
				Op: token.LAND,
				Y:  &ast.CallExpr{Fun: I("containsNull"), Args: []ast.Expr{I("val")}},
			}).WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return1(&ast.CallExpr{
					Fun: Sel(I("errors"), "New"),
					Args: []ast.Expr{&ast.BinaryExpr{
						X:  &ast.BinaryExpr{X: Str("field "), Op: token.ADD, Y: I("field")},
						Op: token.ADD,
						Y:  Str(" cannot be null"),
					}},
				}))))

		bodyBuilder.AddStmt(astbuilder.RangeIndex("field", I("requiredFields")).WithBody(rangeBody))
		g.AddContainsNullIfNeeded()
		g.AddHandlersImport("github.com/go-faster/errors")
	}

	objectFieldsNames := make([]string, 0, len(objectFields))
	for fieldName := range objectFields {
		objectFieldsNames = append(objectFieldsNames, fieldName)
	}
	sort.Strings(objectFieldsNames)

	for _, fieldName := range objectFieldsNames {
		fieldValidationFunc := objectFields[fieldName]
		bodyBuilder.AddStmt(astbuilder.NewAssignBuilder().Lhs(I("val"), I("exists")).Rhs(&ast.IndexExpr{X: I("obj"), Index: Str(fieldName)}))

		ifBody := astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.Assign(I("err"), &ast.CallExpr{Fun: fieldValidationFunc, Args: []ast.Expr{I("val")}})).
			AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return1(&ast.CallExpr{
					Fun:  Sel(I("errors"), "Wrap"),
					Args: []ast.Expr{I("err"), Str("field " + fieldName + " is not valid")},
				}))))

		bodyBuilder.AddStmt(astbuilder.If(&ast.BinaryExpr{
			X:  I("exists"),
			Op: token.LAND,
			Y:  &ast.UnaryExpr{Op: token.NOT, X: &ast.CallExpr{Fun: I("containsNull"), Args: []ast.Expr{I("val")}}},
		}).WithBody(ifBody))

		g.AddContainsNullIfNeeded()
		g.AddHandlersImport("github.com/go-faster/errors")
	}

	bodyBuilder.AddStmt(astbuilder.Return1(I("nil")))

	paramName := "jsonData"
	if len(requiredFields) == 0 && len(objectFields) == 0 {
		paramName = "_"
	}

	fn := astbuilder.Function("Validate" + modelName + "JSON").
		AddParam(astbuilder.SelectorField(paramName, "json", "RawMessage")).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)
	return nil
}

func (g *Generator) AddArrayValidate(modelName string, schema *openapi3.SchemaRef) error {
	const op = "generator.AddArrayValidate"

	elemType, err := g.GetFieldTypeFromSchema(modelName, "Item", schema.Value.Items)
	if err != nil {
		return errors.Wrap(err, op)
	}
	validateFunc := g.GetValidateFuncStmt(elemType, schema.Value.Items.Ref)

	rangeIfBody := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.Assign(I("err"), &ast.CallExpr{Fun: validateFunc, Args: []ast.Expr{I("obj")}})).
		AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.Return1(&ast.CallExpr{
				Fun:  Sel(I("errors"), "Wrapf"),
				Args: []ast.Expr{I("err"), Str("error validating object at index %d"), I("index")},
			}))))

	rangeBody := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.If(&ast.UnaryExpr{
			Op: token.NOT,
			X:  &ast.CallExpr{Fun: I("containsNull"), Args: []ast.Expr{I("obj")}},
		}).WithBody(rangeIfBody))

	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("arr")},
						Type:  &ast.ArrayType{Elt: Sel(I("json"), "RawMessage")},
					},
				},
			},
		}).
		AddStmt(astbuilder.DefineCall("err", Sel(I("json"), "Unmarshal"), I("jsonData"), Amp(I("arr")))).
		AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().AddStmt(astbuilder.Return1(I("err"))))).
		AddStmt(astbuilder.Range("index", "obj", I("arr")).WithBody(rangeBody)).
		AddStmt(astbuilder.Return1(I("nil")))

	fn := astbuilder.Function("Validate" + modelName + "JSON").
		AddParam(astbuilder.SelectorField("jsonData", "json", "RawMessage")).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)
	g.AddHandlersImport("github.com/go-faster/errors")

	return nil
}
