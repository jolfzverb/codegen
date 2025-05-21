package main

import (
	"context"

	"github.com/jolfzverb/codegen/internal/generator"
)

func main() {
	ctx := context.Background()

	err := generator.Generate(ctx, "api/api.yaml", "internal", "internal")
	if err != nil {
		panic(err)
	}
}
