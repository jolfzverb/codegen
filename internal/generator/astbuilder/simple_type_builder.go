package astbuilder

import (
	"go/ast"
)

// SimpleTypeBuilder provides a fluent interface for building simple type expressions
// It can build expressions like "a", "a.B", "package.Type", etc.
type SimpleTypeBuilder struct {
	elements []string
}

// NewSimpleTypeBuilder creates a new SimpleTypeBuilder
func NewSimpleTypeBuilder() *SimpleTypeBuilder {
	return &SimpleTypeBuilder{
		elements: make([]string, 0),
	}
}

// AddElements adds multiple elements to the type expression
// Returns the builder for method chaining
func (stb *SimpleTypeBuilder) AddElements(elements ...string) *SimpleTypeBuilder {
	for _, element := range elements {
		if element != "" {
			stb.elements = append(stb.elements, element)
		}
	}
	return stb
}

func SimpleType(elements ...string) *SimpleTypeBuilder {
	return NewSimpleTypeBuilder().AddElements(elements...)
}

// Build creates the ast.Expr for the simple type
func (stb *SimpleTypeBuilder) Build() ast.Expr {
	if len(stb.elements) == 0 {
		panic("simple type must have at least one element")
	}

	if len(stb.elements) == 1 {
		return ast.NewIdent(stb.elements[0])
	}
	var expr ast.Expr = ast.NewIdent(stb.elements[0])
	for i := 1; i < len(stb.elements); i++ {
		expr = &ast.SelectorExpr{
			X:   expr,
			Sel: ast.NewIdent(stb.elements[i]),
		}
	}
	return expr
}

// Helper methods for common types

// String creates a simple type builder for "string"
func String() *SimpleTypeBuilder {
	return SimpleType("string")
}

// Int creates a simple type builder for "int"
func Int() *SimpleTypeBuilder {
	return SimpleType("int")
}

// Bool creates a simple type builder for "bool"
func Bool() *SimpleTypeBuilder {
	return SimpleType("bool")
}

// Error creates a simple type builder for "error"
func Error() *SimpleTypeBuilder {
	return SimpleType("error")
}

// Context creates a simple type builder for "context.Context"
func Context() *SimpleTypeBuilder {
	return SimpleType("context", "Context")
}

// TypeExpressionBuilder is an interface that can build ast.Expr types
type TypeExpressionBuilder interface {
	Build() ast.Expr
}

// ArrayTypeBuilder provides a fluent interface for building slice types
// It can build expressions like "[]string", "[][]int", etc.
type ArrayTypeBuilder struct {
	element TypeExpressionBuilder
}

// Build creates the ast.Expr for the array type
func (atb *ArrayTypeBuilder) Build() ast.Expr {
	if atb.element == nil {
		panic("array type must have an element type")
	}
	return &ast.ArrayType{
		Elt: atb.element.Build(),
	}
}

// SliceOf creates an ArrayTypeBuilder for []TypeExpressionBuilder
func SliceOf(element TypeExpressionBuilder) *ArrayTypeBuilder {
	if element == nil {
		panic("element cannot be nil")
	}
	return &ArrayTypeBuilder{element: element}
}

// TypeAliasBuilder provides a fluent interface for building type aliases
// It can build declarations like "type NameOfAlias []string"
type TypeAliasBuilder struct {
	name        string
	typeBuilder TypeExpressionBuilder
}

// Build creates the ast.TypeSpec for the type alias
func (tab *TypeAliasBuilder) Build() *ast.TypeSpec {
	if tab.name == "" {
		panic("type alias must have a name")
	}
	if tab.typeBuilder == nil {
		panic("type alias must have a type")
	}

	return &ast.TypeSpec{
		Name: ast.NewIdent(tab.name),
		Type: tab.typeBuilder.Build(),
	}
}


// AliasOf creates a TypeAliasBuilder for "type name Type"
func AliasOf(name string, typeBuilder TypeExpressionBuilder) *TypeAliasBuilder {
	if name == "" {
		panic("alias name cannot be empty")
	}
	if typeBuilder == nil {
		panic("type builder cannot be nil")
	}
	return &TypeAliasBuilder{name: name, typeBuilder: typeBuilder}
}

// MapTypeBuilder provides a fluent interface for building map types like map[K]V
type MapTypeBuilder struct {
	key   TypeExpressionBuilder
	value TypeExpressionBuilder
}

// Build creates the ast.Expr for the map type
func (mtb *MapTypeBuilder) Build() ast.Expr {
	if mtb.key == nil {
		panic("map type must have a key type")
	}
	if mtb.value == nil {
		panic("map type must have a value type")
	}
	return &ast.MapType{
		Key:   mtb.key.Build(),
		Value: mtb.value.Build(),
	}
}

// MapOf creates a MapTypeBuilder for map[K]V
func MapOf(key, value TypeExpressionBuilder) *MapTypeBuilder {
	if key == nil {
		panic("key cannot be nil")
	}
	if value == nil {
		panic("value cannot be nil")
	}
	return &MapTypeBuilder{key: key, value: value}
}
