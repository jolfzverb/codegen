package generator

import (
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

type SchemasFile struct {
	requiredFieldsArePointers bool
	packageImports            []string
	decls                     []*ast.GenDecl
}

type SchemaStruct struct {
	Name   string
	Fields []SchemaField
}

type SchemaField struct {
	Name        string
	Type        string
	TagJSON     []string
	TagValidate []string
	Required    bool
}

func NewSchemasFile(requiredFieldsArePointers bool) *SchemasFile {
	return &SchemasFile{
		requiredFieldsArePointers: requiredFieldsArePointers,
	}
}

func (m *SchemasFile) GenerateImportsSpecs(imp []string) ([]*ast.ImportSpec, []ast.Spec) {
	var systemImports []string //nolint:prealloc
	var libImports []string
	for _, path := range imp {
		prefix := strings.SplitN(path, "/", 2)[0] //nolint:mnd
		if strings.Contains(prefix, ".") {
			libImports = append(libImports, path)

			continue
		}
		systemImports = append(systemImports, path)
	}

	sort.Strings(systemImports)
	sort.Strings(libImports)

	specs := make([]*ast.ImportSpec, 0, len(imp))
	for _, path := range systemImports {
		specs = append(specs, &ast.ImportSpec{Path: Str(path)})
	}

	// Add a space to separate system and library imports
	// but go/ast is too great for that
	for _, path := range libImports {
		specs = append(specs, &ast.ImportSpec{Path: Str(path)})
	}

	declSpecs := make([]ast.Spec, 0, len(specs))
	for _, spec := range specs {
		declSpecs = append(declSpecs, spec)
	}

	return specs, declSpecs
}

func (m *SchemasFile) WriteToOutput(output io.Writer) error {
	const op = "generator.SchemasFile.WriteToOutput"
	// go/ast package is great!
	_, err := output.Write([]byte("// Code generated by github.com/jolfzverb/codegen; DO NOT EDIT.\n\n"))
	if err != nil {
		return errors.Wrap(err, op)
	}

	importSpecs, declSpecs := m.GenerateImportsSpecs(m.packageImports)

	file := &ast.File{
		Name:    ast.NewIdent("models"),
		Imports: importSpecs,
		Decls:   []ast.Decl{},
	}

	if len(declSpecs) > 0 {
		file.Decls = append(file.Decls, &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: declSpecs,
		})
	}

	for _, decl := range m.decls {
		file.Decls = append(file.Decls, decl)
	}

	err = format.Node(output, token.NewFileSet(), file)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (m *SchemasFile) AddSchema(model SchemaStruct) {
	fieldList := make([]*ast.Field, 0, len(model.Fields))
	for _, field := range model.Fields {
		jsonTags := strings.Join(field.TagJSON, ",")
		validateTags := strings.Join(field.TagValidate, ",")

		var tags string
		if len(field.TagJSON) > 0 {
			tags += "json:\"" + jsonTags + "\""
		}
		if len(field.TagValidate) > 0 {
			if len(tags) > 0 {
				tags += " "
			}
			tags += "validate:\"" + validateTags + "\""
		}
		if len(tags) > 0 {
			tags = "`" + tags + "`"
		}
		var typeExpr ast.Expr
		if field.Required {
			typeExpr = ast.NewIdent(field.Type)
		} else {
			typeExpr = Star(ast.NewIdent(field.Type))
		}
		var tag *ast.BasicLit
		if len(tags) > 0 {
			tag = &ast.BasicLit{
				Kind:  token.STRING,
				Value: tags,
			}
		}
		fieldList = append(fieldList, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(field.Name)},
			Type:  typeExpr,
			Tag:   tag,
		})
	}

	m.decls = append(m.decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(model.Name),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fieldList,
					},
				},
			},
		},
	})
}

func (m *SchemasFile) AddTypeAlias(name string, typeName string) {
	m.decls = append(m.decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(name),
				Type: &ast.Ident{
					Name: typeName,
				},
			},
		},
	})
}

func (m *SchemasFile) AddSliceAlias(name string, typeName string) {
	m.decls = append(m.decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(name),
				Type: &ast.ArrayType{
					Elt: ast.NewIdent(typeName),
				},
			},
		},
	})
}

func (m *SchemasFile) AddParamsModel(baseName string, paramType string, params openapi3.Parameters) error {
	const op = "generator.AddParamsModel"
	fields := make([]SchemaField, 0, len(params))
	for _, param := range params {
		name := FormatGoLikeIdentifier(param.Value.Name)
		if !param.Value.Schema.Value.Type.Permits(openapi3.TypeString) {
			return errors.New("only string type parameters are supported for " + paramType + " parameters")
		}
		var jsonTags []string
		var validateTags []string
		jsonTags = append(jsonTags, param.Value.Name)
		if param.Value.Required {
			validateTags = append(validateTags, "required")
		} else {
			jsonTags = append(jsonTags, "omitempty")
			validateTags = append(validateTags, "omitempty")
		}

		validateTags = append(validateTags, GetSchemaValidators(param.Value.Schema)...)
		fieldType, err := m.GetFieldTypeFromSchema(name, "", param.Value.Schema)
		if err != nil {
			return errors.Wrap(err, op)
		}
		required := false
		if !m.requiredFieldsArePointers {
			required = param.Value.Required
		}
		field := SchemaField{
			Name:        name,
			Type:        fieldType,
			TagJSON:     jsonTags,
			TagValidate: validateTags,
			Required:    required,
		}
		fields = append(fields, field)
	}

	model := SchemaStruct{
		Name:   baseName + paramType,
		Fields: fields,
	}
	m.AddSchema(model)

	return nil
}

func (m *SchemasFile) AddHeadersModel(baseName string, headers openapi3.Headers) error {
	const op = "generator.AddHeadersModel"
	fields := make([]SchemaField, 0, len(headers))
	for name, header := range headers {
		if !header.Value.Schema.Value.Type.Permits(openapi3.TypeString) {
			return errors.New("only string type parameters are supported for response headers")
		}
		var jsonTags []string
		var validateTags []string
		jsonTags = append(jsonTags, name)
		if header.Value.Required {
			validateTags = append(validateTags, "required")
		} else {
			jsonTags = append(jsonTags, "omitempty")
			validateTags = append(validateTags, "omitempty")
		}

		validateTags = append(validateTags, GetSchemaValidators(header.Value.Schema)...)
		fieldType, err := m.GetFieldTypeFromSchema(FormatGoLikeIdentifier(name), "", header.Value.Schema)
		if err != nil {
			return errors.Wrap(err, op)
		}
		required := false
		if !m.requiredFieldsArePointers {
			required = header.Value.Required
		}
		field := SchemaField{
			Name:        FormatGoLikeIdentifier(name),
			Type:        fieldType,
			TagJSON:     jsonTags,
			TagValidate: validateTags,
			Required:    required,
		}
		fields = append(fields, field)
	}

	model := SchemaStruct{
		Name:   baseName + "Headers",
		Fields: fields,
	}
	m.AddSchema(model)

	return nil
}

func (m *SchemasFile) GetIntegerType(format string) string {
	integerFormats := map[string]bool{
		"int8": true, "int16": true, "int32": true, "int64": true,
		"uint8": true, "uint16": true, "uint32": true, "uint64": true,
	}
	if ok := integerFormats[format]; ok {
		return format
	}

	return "int"
}

func (m *SchemasFile) AddImport(path string) {
	for _, imp := range m.packageImports {
		if imp == path {
			return
		}
	}
	m.packageImports = append(m.packageImports, path)
}

func (m *SchemasFile) GetStringType(format string) string {
	if format == "date-time" {
		m.AddImport("time")
		return "time.Time"
	}

	return "string"
}

func (m *SchemasFile) GetFieldTypeFromSchema(modelName string, fieldName string,
	fieldSchema *openapi3.SchemaRef,
) (string, error) {
	if fieldSchema.Ref != "" {
		return ParseRefTypeName(fieldSchema.Ref), nil
	}

	var fieldType string
	switch {
	case fieldSchema.Value.Type.Permits(openapi3.TypeString):
		fieldType = m.GetStringType(fieldSchema.Value.Format)
	case fieldSchema.Value.Type.Permits(openapi3.TypeInteger):
		fieldType = m.GetIntegerType(fieldSchema.Value.Format)
	case fieldSchema.Value.Type.Permits(openapi3.TypeNumber):
		fieldType = "float64"
	case fieldSchema.Value.Type.Permits(openapi3.TypeBoolean):
		fieldType = "bool"
	case fieldSchema.Value.Type.Permits(openapi3.TypeObject):
		if fieldSchema.Ref == "" {
			fieldType = modelName + FormatGoLikeIdentifier(fieldName)
		}
	case fieldSchema.Value.Type.Permits(openapi3.TypeArray):
		if fieldSchema.Ref == "" {
			fieldType = modelName + FormatGoLikeIdentifier(fieldName)
		}
	default:
		return "", errors.New("unsupported schema type of field " + fieldName)
	}

	return fieldType, nil
}

func (m *SchemasFile) ProcessObjectSchema(modelName string, schema *openapi3.SchemaRef) error {
	const op = "generator.ProcessObjectSchema"
	model := SchemaStruct{
		Name:   modelName,
		Fields: []SchemaField{},
	}

	requiredFields := make(map[string]bool)
	for _, fieldName := range schema.Value.Required {
		requiredFields[fieldName] = true
	}

	keys := make([]string, 0, len(schema.Value.Properties))
	for key := range schema.Value.Properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, fieldName := range keys {
		fieldSchema := schema.Value.Properties[fieldName]
		var jsonTags []string
		var validateTags []string
		jsonTags = append(jsonTags, fieldName)
		if requiredFields[fieldName] {
			validateTags = append(validateTags, "required")
		} else {
			jsonTags = append(jsonTags, "omitempty")
			validateTags = append(validateTags, "omitempty")
		}

		switch {
		case fieldSchema.Value.Type.Permits(openapi3.TypeObject):
			if fieldSchema.Ref == "" {
				err := m.ProcessObjectSchema(modelName+FormatGoLikeIdentifier(fieldName), fieldSchema)
				if err != nil {
					return errors.Wrap(err, op)
				}
			}
		case fieldSchema.Value.Type.Permits(openapi3.TypeArray):
			if fieldSchema.Ref == "" {
				err := m.ProcessArraySchema(modelName+FormatGoLikeIdentifier(fieldName), fieldSchema)
				if err != nil {
					return errors.Wrap(err, op)
				}
			}
		}

		validateTags = append(validateTags, GetSchemaValidators(fieldSchema)...)

		fieldType, err := m.GetFieldTypeFromSchema(modelName, fieldName, fieldSchema)
		if err != nil {
			return errors.Wrapf(err, op)
		}
		required := false
		if !m.requiredFieldsArePointers {
			required = requiredFields[fieldName]
		}
		field := SchemaField{
			Name:        FormatGoLikeIdentifier(fieldName),
			Type:        fieldType,
			TagJSON:     jsonTags,
			TagValidate: validateTags,
			Required:    required,
		}
		model.Fields = append(model.Fields, field)
	}
	m.AddSchema(model)

	return nil
}

func (m *SchemasFile) ProcessTypeAlias(modelName string, schema *openapi3.SchemaRef) error {
	const op = "generator.ProcessTypeAlias"
	typeName, err := m.GetFieldTypeFromSchema(modelName, "", schema)
	if err != nil {
		return errors.Wrapf(err, op)
	}
	m.AddTypeAlias(modelName, typeName)

	return nil
}

func (m *SchemasFile) ProcessArraySchema(modelName string, schema *openapi3.SchemaRef,
) error {
	const op = "generator.ProcessArraySchema"
	var elemType string

	if schema.Value.Items.Ref == "" {
		itemsSchema := schema.Value.Items
		switch {
		case itemsSchema.Value.Type.Permits(openapi3.TypeObject):
			err := m.ProcessObjectSchema(modelName+"Item", itemsSchema)
			if err != nil {
				return errors.Wrap(err, op)
			}
		case itemsSchema.Value.Type.Permits(openapi3.TypeArray):
			err := m.ProcessArraySchema(modelName+"Item", itemsSchema)
			if err != nil {
				return errors.Wrap(err, op)
			}
		}
	}

	elemType, err := m.GetFieldTypeFromSchema(modelName, "Item", schema.Value.Items)
	if err != nil {
		return errors.Wrapf(err, op)
	}

	m.AddSliceAlias(modelName, elemType)

	return nil
}

func (m *SchemasFile) ProcessSchema(modelName string, schema *openapi3.SchemaRef) error {
	const op = "generator.ProcessSchema"
	switch {
	case schema.Value.Type.Permits(openapi3.TypeObject):
		err := m.ProcessObjectSchema(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeArray):
		err := m.ProcessArraySchema(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeString):
		err := m.ProcessTypeAlias(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeBoolean):
		err := m.ProcessTypeAlias(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeInteger):
		err := m.ProcessTypeAlias(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeNumber):
		err := m.ProcessTypeAlias(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	}

	return errors.Errorf("unsupported schema type %s for model %s", schema.Value.Type, modelName)
}

func (m *SchemasFile) GenerateRequestModel(baseName string, contentType string, pathParams openapi3.Parameters,
	queryParams openapi3.Parameters, headers openapi3.Parameters, cookieParams openapi3.Parameters,
	body *openapi3.RequestBodyRef,
) {
	model := SchemaStruct{
		Name:   baseName + "Request",
		Fields: []SchemaField{},
	}
	if len(pathParams) > 0 {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Path",
			Type:        baseName + "PathParams",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    true,
		})
	}
	if len(queryParams) > 0 {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Query",
			Type:        baseName + "QueryParams",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    true,
		})
	}
	if len(headers) > 0 {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Headers",
			Type:        baseName + "Headers",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    true,
		})
	}
	if len(cookieParams) > 0 {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Cookies",
			Type:        baseName + "Cookies",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    true,
		})
	}
	if body != nil && body.Value != nil {
		content, ok := body.Value.Content[contentType]
		if ok && content.Schema != nil {
			typeName := baseName + "RequestBody"

			if content.Schema.Ref != "" {
				typeName = ParseRefTypeName(content.Schema.Ref)
			}

			model.Fields = append(model.Fields, SchemaField{
				Name:        "Body",
				Type:        typeName,
				TagJSON:     []string{},
				TagValidate: []string{},
				Required:    body.Value.Required,
			})
		}
	}

	m.AddSchema(model)
}
