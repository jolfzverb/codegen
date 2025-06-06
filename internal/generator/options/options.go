package options

import (
	"flag"

	"github.com/go-faster/errors"
)

type Options struct {
	PackagePrefix             string
	DirPrefix                 string
	YAMLFileName              string
	RequiredFieldsArePointers bool
}

func GetOptions() (*Options, error) {
	opts := Options{}

	flag.StringVar(&opts.YAMLFileName, "f", "", "Path to the OpenAPI YAML file")
	flag.StringVar(&opts.DirPrefix, "d", "internal", "Directory prefix for generated files")
	flag.StringVar(&opts.PackagePrefix, "p", "internal", "Package prefix for imports")
	flag.BoolVar(&opts.RequiredFieldsArePointers, "pointers", false, "Generate required fields as pointers")

	flag.Parse()

	if opts.YAMLFileName == "" {
		return nil, errors.New("-f flag is required")
	}

	return &opts, nil
}
