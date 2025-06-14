package generator

import (
	"context"
	"io"
	"os"
	"path"
	"strings"

	"github.com/go-faster/errors"
	"github.com/jolfzverb/codegen/internal/generator/options"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const directoryPermissions = 0o755

func GetModelName(yamlFilePath string) string {
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

func Generate(ctx context.Context, opts *options.Options) error {
	const op = "generator.Generate"
	file, err := os.Open(opts.YAMLFileName)
	if err != nil {
		return errors.Wrap(err, op)
	}
	defer file.Close()

	reader := io.Reader(file)

	modelName := GetModelName(opts.YAMLFileName)

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

	err = GenerateToIO(ctx, reader, schemasOutput, handlerOutput,
		path.Join(opts.PackagePrefix, "generated", modelName), modelName, opts)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}
