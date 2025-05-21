package generator

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
)

func GetSchemaValidators(schema *openapi3.SchemaRef) []string {
	var validateTags []string
	switch {
	case schema.Value.Type.Permits(openapi3.TypeString):
		if schema.Value.MinLength > 0 {
			validateTags = append(validateTags, "min="+strconv.FormatUint(schema.Value.MinLength, 10))
		}
		if schema.Value.MaxLength != nil {
			validateTags = append(validateTags, "max="+strconv.FormatUint(*schema.Value.MaxLength, 10))
		}
		if schema.Value.Pattern != "" {
			slog.Warn("pattern validator is not supported", slog.String("pattern", schema.Value.Pattern))
		}

	case schema.Value.Type.Permits(openapi3.TypeInteger):
		if schema.Value.Min != nil {
			validateTags = append(validateTags, "min="+fmt.Sprint(*schema.Value.Min))
		}
		if schema.Value.Max != nil {
			validateTags = append(validateTags, "max="+fmt.Sprint(*schema.Value.Max))
		}
		if schema.Value.MultipleOf != nil {
			slog.Warn("multipleOf validator is not supported")
		}
		if schema.Value.ExclusiveMax {
			slog.Warn("exclusiveMax validator is not supported")
		}
		if schema.Value.ExclusiveMin {
			slog.Warn("exclusiveMin validator is not supported")
		}

	case schema.Value.Type.Permits(openapi3.TypeNumber):
		if schema.Value.Min != nil {
			validateTags = append(validateTags, "min="+fmt.Sprint(*schema.Value.Min))
		}
		if schema.Value.Max != nil {
			validateTags = append(validateTags, "max="+fmt.Sprint(*schema.Value.Max))
		}
		if schema.Value.MultipleOf != nil {
			slog.Warn("multipleOf validator is not supported")
		}
		if schema.Value.ExclusiveMax {
			slog.Warn("exclusiveMax validator is not supported")
		}
		if schema.Value.ExclusiveMin {
			slog.Warn("exclusiveMin validator is not supported")
		}

	case schema.Value.Type.Permits(openapi3.TypeArray):
		if schema.Value.MinItems > 0 {
			validateTags = append(validateTags, "min="+strconv.FormatUint(schema.Value.MinItems, 10))
		}
		if schema.Value.MaxItems != nil {
			validateTags = append(validateTags, "max="+strconv.FormatUint(*schema.Value.MaxItems, 10))
		}
		if schema.Value.UniqueItems {
			validateTags = append(validateTags, "unique")
		}
		itemsValidators := GetSchemaValidators(schema.Value.Items)
		if len(itemsValidators) > 0 {
			validateTags = append(validateTags, "dive")
			validateTags = append(validateTags, itemsValidators...)
		}
	}

	return validateTags
}
