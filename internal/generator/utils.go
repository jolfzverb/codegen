package generator

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func GetSchemaValidators(schema *openapi3.SchemaRef) []string {
	var validateTags []string
	switch {
	case schema.Value.Type.Permits(openapi3.TypeString):
		if schema.Value.MinLength > 0 {
			validateTags = append(validateTags, "min="+fmt.Sprint(schema.Value.MinLength))
		}
		if schema.Value.MaxLength != nil {
			validateTags = append(validateTags, "max="+fmt.Sprint(*schema.Value.MaxLength))
		}
		if schema.Value.Pattern != "" {
			fmt.Printf("Warn: pattern validator(%s) is not supported\n", schema.Value.Pattern)
		}

	case schema.Value.Type.Permits(openapi3.TypeInteger):
		if schema.Value.Min != nil {
			validateTags = append(validateTags, "min="+fmt.Sprint(*schema.Value.Min))
		}
		if schema.Value.Max != nil {
			validateTags = append(validateTags, "max="+fmt.Sprint(*schema.Value.Max))
		}
		if schema.Value.MultipleOf != nil {
			fmt.Printf("Warn: multipleOf validator is not supported\n")
		}
		if schema.Value.ExclusiveMax {
			fmt.Printf("Warn: exclusiveMax validator is not supported\n")
		}
		if schema.Value.ExclusiveMin {
			fmt.Printf("Warn: exclusiveMin validator is not supported\n")
		}

	case schema.Value.Type.Permits(openapi3.TypeNumber):
		if schema.Value.Min != nil {
			validateTags = append(validateTags, "min="+fmt.Sprint(*schema.Value.Min))
		}
		if schema.Value.Max != nil {
			validateTags = append(validateTags, "max="+fmt.Sprint(*schema.Value.Max))
		}
		if schema.Value.MultipleOf != nil {
			fmt.Printf("Warn: multipleOf validator is not supported\n")
		}
		if schema.Value.ExclusiveMax {
			fmt.Printf("Warn: exclusiveMax validator is not supported\n")
		}
		if schema.Value.ExclusiveMin {
			fmt.Printf("Warn: exclusiveMin validator is not supported\n")
		}

	case schema.Value.Type.Permits(openapi3.TypeArray):
		if schema.Value.MinItems > 0 {
			validateTags = append(validateTags, "min="+fmt.Sprint(schema.Value.MinItems))
		}
		if schema.Value.MaxItems != nil {
			validateTags = append(validateTags, "max="+fmt.Sprint(*schema.Value.MaxItems))
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
