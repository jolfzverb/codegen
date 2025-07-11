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
	yaml         *openapi3.T

	// strings
	PackageName      string
	ImportPrefix     string
	ModelsImportPath string
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

func (g *Generator) Gen() {
	const op = "generator.Generate"

	// one time
	g.InitHandlerFields(g.PackageName, g.ModelsImportPath)

	if g.yaml.Paths != nil && len(g.yaml.Paths.Map()) > 0 {
		err := g.ProcessPaths(g.yaml.Paths)
		if err != nil {
			panic(errors.Wrap(err, op))
		}
	}

	if g.yaml.Components != nil && g.yaml.Components.Schemas != nil {
		err := g.ProcessSchemas(g.yaml.Components.Schemas)
		if err != nil {
			panic(errors.Wrap(err, op))
		}
	}

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

func (g *Generator) PrepareAndRead(reader io.Reader) error {
	const op = "generator.PrepareAndRead"
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	var err error
	g.yaml, err = loader.LoadFromIoReader(reader)
	if err != nil {
		return errors.Wrap(err, op)
	}
	err = g.yaml.Validate(ctx)
	if err != nil {
		return errors.Wrap(err, op)
	}

	g.NewSchemasFile()
	g.NewHandlersFile()

	return nil
}

func (g *Generator) PrepareFiles() error {
	const op = "generator.PrepareFiles"

	file, err := os.Open(g.Opts.YAMLFileName)
	if err != nil {
		return errors.Wrap(err, op)
	}
	defer file.Close()

	reader := io.Reader(file)

	g.PackageName = g.GetModelName(g.Opts.YAMLFileName)

	handlersPath := path.Join(g.Opts.DirPrefix, "generated", g.PackageName)
	schemasPath := path.Join(handlersPath, "models")
	err = os.MkdirAll(schemasPath, directoryPermissions)
	if err != nil {
		return errors.Wrap(err, op)
	}

	g.ImportPrefix = path.Join(g.Opts.PackagePrefix, "generated", g.PackageName)
	g.ModelsImportPath = path.Join(g.ImportPrefix, "models")
	err = g.PrepareAndRead(reader)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (g *Generator) GenerateFiles() error {
	g.Gen()
	return nil
}
func (g *Generator) WriteOutFiles() error {
	const op = "generator.WriteOutFiles"

	modelName := g.GetModelName(g.Opts.YAMLFileName)
	handlersPath := path.Join(g.Opts.DirPrefix, "generated", modelName)
	schemasPath := path.Join(handlersPath, "models")
	schemasOutput, err := os.Create(path.Join(schemasPath, "models.go"))
	if err != nil {
		return errors.Wrap(err, op)
	}
	defer schemasOutput.Close()

	handlersOutput, err := os.Create(path.Join(handlersPath, "handlers.go"))
	if err != nil {
		return errors.Wrap(err, op)
	}
	defer handlersOutput.Close()

	err = g.WriteToOutput(schemasOutput, handlersOutput)
	if err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}

func (g *Generator) Generate(ctx context.Context) error {
	const op = "generator.Generate"

	err := g.PrepareFiles()
	if err != nil {
		return errors.Wrap(err, op)
	}
	err = g.GenerateFiles()
	if err != nil {
		return errors.Wrap(err, op)
	}
	err = g.WriteOutFiles()
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}
