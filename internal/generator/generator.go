package generator

import (
	"context"
	"io"
	"path"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

type Generator struct {
	SchemasFile  *SchemasFile
	HandlersFile *HandlersFile
}

func NewGenerator(importPrefix string, packageName string) *Generator {
	return &Generator{
		SchemasFile:  NewSchemasFile(),
		HandlersFile: NewHandlersFile(packageName, importPrefix, path.Join(importPrefix, "models")),
	}
}

func (g *Generator) WriteToOutput(modelsOutput io.Writer, handlersOutput io.Writer) error {
	const op = "generator.Generator.WriteToOutput"
	err := g.SchemasFile.WriteToOutput(modelsOutput)
	if err != nil {
		return errors.Wrap(err, op)
	}
	err = g.HandlersFile.WriteToOutput(handlersOutput)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (g *Generator) Generate(yaml *openapi3.T) {
	const op = "generator.Generate"
	if yaml.Components != nil && yaml.Components.Schemas != nil {
		err := g.ProcessSchemas(yaml.Components.Schemas)
		if err != nil {
			panic(errors.Wrap(err, op))
		}
	}
	if yaml.Paths != nil && len(yaml.Paths.Map()) > 0 {
		err := g.ProcessPaths(yaml.Paths)
		if err != nil {
			panic(errors.Wrap(err, op))
		}
	}
}

func GenerateToIO(ctx context.Context, input io.Reader, schemasOutput io.Writer, handlersOutput io.Writer,
	importPrefix string, packageName string,
) error {
	const op = "generator.GenerateToIO"
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	yaml, err := loader.LoadFromIoReader(input)
	if err != nil {
		return errors.Wrap(err, op)
	}
	err = yaml.Validate(ctx)
	if err != nil {
		return errors.Wrap(err, op)
	}
	generator := NewGenerator(importPrefix, packageName)
	generator.Generate(yaml)

	err = generator.WriteToOutput(schemasOutput, handlersOutput)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}
