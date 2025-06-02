package main

import (
	"context"

	"github.com/jolfzverb/codegen/internal/generator"
	"github.com/jolfzverb/codegen/internal/generator/options"
)

func main() {
	opts, err := options.GetOptions()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	err = generator.Generate(ctx, opts)
	if err != nil {
		panic(err)
	}
}
