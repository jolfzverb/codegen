package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/jolfzverb/codegen/internal/generator"
)

func main() {
	yamlPath := flag.String("f", "", "Path to the OpenAPI YAML file")
	dirPrefix := flag.String("d", "internal", "Directory prefix for generated files")
	pkgPrefix := flag.String("p", "github.com/jolfzverb/codegen/internal", "Package prefix for imports")

	flag.Parse()

	if *yamlPath == "" {
		slog.Error("-yaml flag is required")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	err := generator.Generate(ctx, *yamlPath, *dirPrefix, *pkgPrefix)
	if err != nil {
		panic(err)
	}
}
