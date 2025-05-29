package generator

import (
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

func (g *Generator) ProcessObjectSchema(modelName string, schema *openapi3.SchemaRef) error {
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

		var fieldType string
		switch {
		case fieldSchema.Value.Type.Permits(openapi3.TypeString):
			fieldType = "string"
		case fieldSchema.Value.Type.Permits(openapi3.TypeInteger):
			fieldType = "int"
		case fieldSchema.Value.Type.Permits(openapi3.TypeNumber):
			fieldType = "float64"
		case fieldSchema.Value.Type.Permits(openapi3.TypeBoolean):
			fieldType = "bool"
		case fieldSchema.Value.Type.Permits(openapi3.TypeObject):
			if fieldSchema.Ref == "" {
				err := g.ProcessObjectSchema(modelName+FormatGoLikeIdentifier(fieldName), fieldSchema)
				if err != nil {
					return errors.Wrap(err, op)
				}
				fieldType = modelName + FormatGoLikeIdentifier(fieldName)
			}
		case fieldSchema.Value.Type.Permits(openapi3.TypeArray):
			if fieldSchema.Ref == "" {
				err := g.ProcessArraySchema(modelName+FormatGoLikeIdentifier(fieldName), fieldSchema)
				if err != nil {
					return errors.Wrap(err, op)
				}
				fieldType = modelName + FormatGoLikeIdentifier(fieldName)
			}
		default:
			fieldType = "invalid"
		}

		validateTags = append(validateTags, GetSchemaValidators(fieldSchema)...)

		if fieldSchema.Ref != "" {
			fieldType = ParseRefTypeName(fieldSchema.Ref)
		}
		field := SchemaField{
			Name:        FormatGoLikeIdentifier(fieldName),
			Type:        fieldType,
			TagJSON:     jsonTags,
			TagValidate: validateTags,
		}
		model.Fields = append(model.Fields, field)
	}
	g.SchemasFile.AddSchema(model)

	return nil
}

func (g *Generator) ProcessTypeAlias(modelName string, typeName string) error {
	g.SchemasFile.AddTypeAlias(modelName, typeName)

	return nil
}

func (g *Generator) ProcessArraySchema(modelName string, schema *openapi3.SchemaRef,
) error {
	const op = "generator.ProcessArraySchema"

	if schema.Value.Items.Ref != "" {
		g.SchemasFile.AddSliceAlias(modelName, ParseRefTypeName(schema.Value.Items.Ref))
		return nil
	}

	itemsSchema := schema.Value.Items
	switch {
	case itemsSchema.Value.Type.Permits(openapi3.TypeString):
		g.SchemasFile.AddSliceAlias(modelName, "string")

		return nil
	case itemsSchema.Value.Type.Permits(openapi3.TypeBoolean):
		g.SchemasFile.AddSliceAlias(modelName, "bool")

		return nil
	case itemsSchema.Value.Type.Permits(openapi3.TypeInteger):
		g.SchemasFile.AddSliceAlias(modelName, "int")

		return nil
	case itemsSchema.Value.Type.Permits(openapi3.TypeNumber):
		g.SchemasFile.AddSliceAlias(modelName, "float64")

		return nil
	case itemsSchema.Value.Type.Permits(openapi3.TypeObject):
		err := g.ProcessObjectSchema(modelName+"Item", itemsSchema)
		if err != nil {
			return errors.Wrap(err, op)
		}
		g.SchemasFile.AddSliceAlias(modelName, modelName+"Item")

		return nil
	case itemsSchema.Value.Type.Permits(openapi3.TypeArray):
		err := g.ProcessArraySchema(modelName+"Item", itemsSchema)
		if err != nil {
			return errors.Wrap(err, op)
		}
		g.SchemasFile.AddSliceAlias(modelName, modelName+"Item")

		return nil
	}

	return errors.Errorf("unsupported schema type %s for model %s", itemsSchema.Value.Type, modelName)
}

func (g *Generator) ProcessSchema(modelName string, schema *openapi3.SchemaRef) error {
	const op = "generator.ProcessSchema"
	switch {
	case schema.Value.Type.Permits(openapi3.TypeObject):
		err := g.ProcessObjectSchema(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeArray):
		err := g.ProcessArraySchema(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeString):
		err := g.ProcessTypeAlias(modelName, "string")
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeBoolean):
		err := g.ProcessTypeAlias(modelName, "bool")
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeInteger):
		err := g.ProcessTypeAlias(modelName, "int")
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	case schema.Value.Type.Permits(openapi3.TypeNumber):
		err := g.ProcessTypeAlias(modelName, "float64")
		if err != nil {
			return errors.Wrap(err, op)
		}

		return nil
	}

	return errors.Errorf("unsupported schema type %s for model %s", schema.Value.Type, modelName)
}

func (g *Generator) ProcessSchemas(schemas map[string]*openapi3.SchemaRef) error {
	const op = "generator.ProcessSchemas"
	modelKeys := make([]string, 0, len(schemas))
	for modelName := range schemas {
		modelKeys = append(modelKeys, modelName)
	}
	sort.Strings(modelKeys)

	for _, modelName := range modelKeys {
		schema := schemas[modelName]
		err := g.ProcessSchema(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}
	}

	return nil
}

func (g *Generator) GenerateRequestModel(baseName string, params DetectedParams) {
	model := SchemaStruct{
		Name:   baseName + "Request",
		Fields: []SchemaField{},
	}
	if params.HasPath {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Path",
			Type:        baseName + "PathParams",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    true,
		})
	}
	if params.HasQuery {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Query",
			Type:        baseName + "QueryParams",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    true,
		})
	}
	if params.HasHeaders {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Headers",
			Type:        baseName + "Headers",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    true,
		})
	}
	if params.HasCookies {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Cookies",
			Type:        baseName + "Cookies",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    true,
		})
	}
	if params.HasRequestBody {
		model.Fields = append(model.Fields, SchemaField{
			Name:        "Body",
			Type:        baseName + "RequestBody",
			TagJSON:     []string{},
			TagValidate: []string{},
			Required:    params.BodyRequired,
		})
	}

	g.SchemasFile.AddSchema(model)
}
