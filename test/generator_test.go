package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jolfzverb/codegen/internal/generator"
	"github.com/jolfzverb/codegen/internal/generator/options"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
)

func TestGeneratorCreatesFiles(t *testing.T) {
	tmpDir := t.TempDir()

	opts := &options.Options{
		DirPrefix:     tmpDir,
		PackagePrefix: "github.com/jolfzverb/codegen/internal/generated",
		YAMLFiles: []string{
			"yamls/api.yaml",
			"yamls/api2.yaml",
			"yamls/def.yaml",
		},
		RequiredFieldsArePointers: false,
	}

	// Run the generator
	ctx := context.Background()
	gen := generator.NewGenerator(opts)
	err := gen.Generate(ctx)
	assert.NoError(t, err)

	// Check that files are created
	expectedFiles := []string{
		"generated/api/handlers.go",
		"generated/api/apimodels/models.go",
		"generated/api2/handlers.go",
		"generated/api2/api2models/models.go",
		"generated/def/handlers.go",
		"generated/def/defmodels/models.go",
	}
	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		_, err := os.Stat(fullPath)
		assert.NoError(t, err, "File should exist: %s", fullPath)

		// Verify the content using goldie
		content, err := os.ReadFile(fullPath)
		assert.NoError(t, err)
		g := goldie.New(t, goldie.WithNameSuffix(""))
		g.Assert(t, file, content)
	}
}
