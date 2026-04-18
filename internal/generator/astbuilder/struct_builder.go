package astbuilder

import (
	"go/ast"
)

// StructBuilder provides a fluent interface for building Go structs
type StructBuilder struct {
	name   string
	fields []*FieldBuilder
}

// NewStructBuilder creates a new StructBuilder
func NewStructBuilder() *StructBuilder {
	return &StructBuilder{
		fields: make([]*FieldBuilder, 0),
	}
}

// Struct creates a StructBuilder with name and optional fields
func Struct(name string, fields ...*FieldBuilder) *StructBuilder {
	sb := &StructBuilder{
		name:   name,
		fields: make([]*FieldBuilder, 0, len(fields)),
	}
	for _, f := range fields {
		if f == nil {
			panic("field builder cannot be nil")
		}
		sb.fields = append(sb.fields, f)
	}
	return sb
}

// WithName sets the struct name
// Returns the builder for method chaining
func (sb *StructBuilder) WithName(name string) *StructBuilder {
	sb.name = name
	return sb
}

// AddFields adds multiple fields to the struct
// Returns the builder for method chaining
func (sb *StructBuilder) AddFields(fieldBuilders ...*FieldBuilder) *StructBuilder {
	for _, fieldBuilder := range fieldBuilders {
		if fieldBuilder == nil {
			panic("field builder cannot be nil")
		}
		sb.fields = append(sb.fields, fieldBuilder)
	}
	return sb
}

// Build creates the ast.TypeSpec for the struct
func (sb *StructBuilder) Build() *ast.TypeSpec {
	if sb.name == "" {
		panic("struct must have a name")
	}

	astFields := make([]*ast.Field, len(sb.fields))
	for i, fieldBuilder := range sb.fields {
		astFields[i] = fieldBuilder.Build()
	}

	return &ast.TypeSpec{
		Name: ast.NewIdent(sb.name),
		Type: &ast.StructType{
			Fields: &ast.FieldList{
				List: astFields,
			},
		},
	}
}
