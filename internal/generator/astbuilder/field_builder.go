package astbuilder

import (
	"go/ast"
	"go/token"
	"strings"
)

// FieldBuilder provides a fluent interface for building ast.Field
type FieldBuilder struct {
	name         string
	typeBuilder  TypeExpressionBuilder
	jsonTags     []string
	validateTags []string
}

// NewFieldBuilder creates a new FieldBuilder
func NewFieldBuilder() *FieldBuilder {
	return &FieldBuilder{
		typeBuilder:  SimpleType(),
		jsonTags:     make([]string, 0),
		validateTags: make([]string, 0),
	}
}

// Field creates a named field with the given type
func Field(name string, typeBuilder TypeExpressionBuilder) *FieldBuilder {
	return NewFieldBuilder().WithName(name).WithType(typeBuilder)
}

// WithName sets the field name
// Returns the builder for method chaining
func (fb *FieldBuilder) WithName(name string) *FieldBuilder {
	fb.name = name
	return fb
}

// WithType sets the field type using a TypeExpressionBuilder (SimpleTypeBuilder or ArrayTypeBuilder)
// Returns the builder for method chaining
func (fb *FieldBuilder) WithType(typeBuilder TypeExpressionBuilder) *FieldBuilder {
	if typeBuilder == nil {
		panic("type builder cannot be nil")
	}
	fb.typeBuilder = typeBuilder
	return fb
}

// AddJSONTags adds multiple JSON tags to the field
// Returns the builder for method chaining
func (fb *FieldBuilder) AddJSONTags(tags ...string) *FieldBuilder {
	for _, tag := range tags {
		if tag != "" {
			fb.jsonTags = append(fb.jsonTags, tag)
		}
	}
	return fb
}

// AddValidateTags adds multiple validate tags to the field
// Returns the builder for method chaining
func (fb *FieldBuilder) AddValidateTags(tags ...string) *FieldBuilder {
	for _, tag := range tags {
		if tag != "" {
			fb.validateTags = append(fb.validateTags, tag)
		}
	}
	return fb
}

// Build creates the ast.Field
func (fb *FieldBuilder) Build() *ast.Field {
	if fb.typeBuilder == nil {
		panic("field must have a type builder")
	}

	field := &ast.Field{
		Type: fb.typeBuilder.Build(),
	}

	if fb.name != "" {
		field.Names = []*ast.Ident{ast.NewIdent(fb.name)}
	}

	tagParts := make([]string, 0)

	if len(fb.jsonTags) > 0 {
		jsonTag := strings.Join(fb.jsonTags, ",")
		tagParts = append(tagParts, "json:\""+jsonTag+"\"")
	}

	if len(fb.validateTags) > 0 {
		validateTag := strings.Join(fb.validateTags, ",")
		tagParts = append(tagParts, "validate:\""+validateTag+"\"")
	}

	if len(tagParts) > 0 {
		fullTag := strings.Join(tagParts, " ")
		field.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`" + fullTag + "`",
		}
	}

	return field
}

// Helper methods for common field types

// StringField creates a field builder for a string field
func StringField(name string) *FieldBuilder {
	return Field(name, String())
}

// IntField creates a field builder for an int field
func IntField(name string) *FieldBuilder {
	return Field(name, Int())
}

// BoolField creates a field builder for a bool field
func BoolField(name string) *FieldBuilder {
	return Field(name, Bool())
}

// ErrorField creates a field builder for an error field (unnamed)
func ErrorField() *FieldBuilder {
	return NewFieldBuilder().WithType(Error())
}

// ContextField creates a field builder for a context.Context field
func ContextField(name string) *FieldBuilder {
	return Field(name, Context())
}
