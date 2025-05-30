package generator

import (
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

func (g *Generator) ProcessSchemas(schemas map[string]*openapi3.SchemaRef) error {
	const op = "generator.ProcessSchemas"
	modelKeys := make([]string, 0, len(schemas))
	for modelName := range schemas {
		modelKeys = append(modelKeys, modelName)
	}
	sort.Strings(modelKeys)

	for _, modelName := range modelKeys {
		schema := schemas[modelName]
		err := g.SchemasFile.ProcessSchema(modelName, schema)
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
