package generator

import (
	"context"
	"io"
	"os"
	"path"
	"strings"

	"github.com/go-faster/errors"
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

	lowerCaser := cases.Lower(language.Und)

	return lowerCaser.String(fileName)
}

func Generate(ctx context.Context, yamlFilePath string, outputPathPrefix string, outputImportPrefix string) error {
	const op = "generator.Generate"
	file, err := os.Open(yamlFilePath)
	if err != nil {
		return errors.Wrap(err, op)
	}
	defer file.Close()

	reader := io.Reader(file)

	modelName := GetModelName(yamlFilePath)
	handlersPath := path.Join(outputPathPrefix, "generated", modelName)
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
		path.Join(outputImportPrefix, "generated", modelName), modelName)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}
