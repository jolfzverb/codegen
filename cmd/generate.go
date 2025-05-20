package main

import (
	"context"

	"github.com/jolfzverb/codegen/internal/generator"
)

func main() {
	ctx := context.Background()

	generator.Generate(ctx, "api/api.yaml", "internal", "internal")
}
