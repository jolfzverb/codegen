package options

import (
	"flag"

	"github.com/go-faster/errors"
)

type Options struct {
	PackagePrefix             string
	DirPrefix                 string
	YAMLFiles                 []string
	RequiredFieldsArePointers bool
}

func GetOptions() (*Options, error) {
	opts := Options{}

	flag.StringVar(&opts.DirPrefix, "d", "internal", "Directory prefix for generated files")
	flag.StringVar(&opts.PackagePrefix, "p", "internal", "Package prefix for imports")
	flag.BoolVar(&opts.RequiredFieldsArePointers, "pointers", false, "Generate required fields as pointers")

	flag.Parse()
	opts.YAMLFiles = flag.Args()

	if len(opts.YAMLFiles) == 0 {
		return nil, errors.New("at least one file must be provided")
	}

	return &opts, nil
}
