package generator

import (
	"context"
	"io"
	"os"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
	"github.com/jolfzverb/codegen/internal/generator/options"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const directoryPermissions = 0o755

type Generator struct {
	Opts *options.Options

	SchemasFile  *SchemasFile
	HandlersFile *HandlersFile
}

func NewGenerator(opts *options.Options) *Generator {
	return &Generator{
		Opts: opts,
	}
}

func (g *Generator) WriteToOutput(modelsOutput io.Writer, handlersOutput io.Writer) error {
	const op = "generator.Generator.WriteToOutput"
	err := g.WriteSchemasToOutput(modelsOutput)
	if err != nil {
		return errors.Wrap(err, op)
	}
	err = g.WriteHandlersToOutput(handlersOutput)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (g *Generator) Gen(yaml *openapi3.T) {
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

func (g *Generator) GenerateToIO(ctx context.Context, input io.Reader, schemasOutput io.Writer, handlersOutput io.Writer,
	importPrefix string, packageName string, opts *options.Options,
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

	g.SchemasFile = g.NewSchemasFile(opts.RequiredFieldsArePointers)
	g.HandlersFile = g.NewHandlersFile(packageName, importPrefix, path.Join(importPrefix, "models"), opts.RequiredFieldsArePointers)
	g.Gen(yaml)

	err = g.WriteToOutput(schemasOutput, handlersOutput)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (g *Generator) GetModelName(yamlFilePath string) string {
	parts := strings.Split(yamlFilePath, "/")
	if len(parts) == 0 {
		return ""
	}
	fileName := parts[len(parts)-1]
	fileName = strings.TrimSuffix(fileName, ".yaml")
	fileName = strings.TrimSuffix(fileName, ".yml")
	fileName = strings.ReplaceAll(fileName, "_", "")
	fileName = strings.ReplaceAll(fileName, "-", "")

	lowerCaser := cases.Lower(language.Und)

	return lowerCaser.String(fileName)
}

func (g *Generator) Generate(ctx context.Context, opts *options.Options) error {
	const op = "generator.Generate"
	file, err := os.Open(opts.YAMLFileName)
	if err != nil {
		return errors.Wrap(err, op)
	}
	defer file.Close()

	reader := io.Reader(file)

	modelName := g.GetModelName(opts.YAMLFileName)

	handlersPath := path.Join(opts.DirPrefix, "generated", modelName)
	schemasPath := path.Join(handlersPath, "models")
	err = os.MkdirAll(schemasPath, directoryPermissions)
	if err != nil {
		return errors.Wrap(err, op)
	}

	schemasOutput, err := os.Create(path.Join(schemasPath, "models.go"))
	if err != nil {
		return errors.Wrap(err, op)
	}
	defer schemasOutput.Close()

	handlerOutput, err := os.Create(path.Join(handlersPath, "handlers.go"))
	if err != nil {
		return errors.Wrap(err, op)
	}
	defer handlerOutput.Close()

	err = g.GenerateToIO(ctx, reader, schemasOutput, handlerOutput,
		path.Join(opts.PackagePrefix, "generated", modelName), modelName, opts)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}
