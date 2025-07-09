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
		err := g.ProcessSchema(modelName, schema)
		if err != nil {
			return errors.Wrap(err, op)
		}
	}

	return nil
}
