package astbuilder

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestNewSimpleTypeBuilder(t *testing.T) {
	builder := NewSimpleTypeBuilder()

	if builder == nil {
		t.Fatal("NewSimpleTypeBuilder returned nil")
	}

	if builder.elements == nil {
		t.Fatal("elements slice is nil")
	}

	if len(builder.elements) != 0 {
		t.Errorf("Expected empty elements initially, got %d", len(builder.elements))
	}
}

func TestSimpleTypeBuilder_AddElements(t *testing.T) {
	builder := NewSimpleTypeBuilder()

	result := builder.AddElements("context", "Context")
	if result != builder {
		t.Error("AddElements should return the builder for chaining")
	}

	if len(builder.elements) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(builder.elements))
	}

	if builder.elements[0] != "context" {
		t.Errorf("Expected first element 'context', got %s", builder.elements[0])
	}

	if builder.elements[1] != "Context" {
		t.Errorf("Expected second element 'Context', got %s", builder.elements[1])
	}

	// Empty strings should be ignored
	builder.AddElements("", "Type", "")
	if len(builder.elements) != 3 {
		t.Errorf("Expected 3 elements after adding with empty strings, got %d", len(builder.elements))
	}

	if builder.elements[2] != "Type" {
		t.Errorf("Expected third element 'Type', got %s", builder.elements[2])
	}
}

func TestSimpleTypeBuilder_Build(t *testing.T) {
	// Single element -> ast.Ident
	builder := NewSimpleTypeBuilder().AddElements("string")
	expr := builder.Build()

	if ident, ok := expr.(*ast.Ident); ok {
		if ident.Name != "string" {
			t.Errorf("Expected ident name 'string', got %s", ident.Name)
		}
	} else {
		t.Error("Single element should create ast.Ident")
	}

	// Two elements -> ast.SelectorExpr
	builder = NewSimpleTypeBuilder().AddElements("context", "Context")
	expr = builder.Build()

	if selector, ok := expr.(*ast.SelectorExpr); ok {
		if ident, ok := selector.X.(*ast.Ident); ok {
			if ident.Name != "context" {
				t.Errorf("Expected selector X name 'context', got %s", ident.Name)
			}
		} else {
			t.Error("Selector X should be ast.Ident")
		}

		if selector.Sel.Name != "Context" {
			t.Errorf("Expected selector Sel name 'Context', got %s", selector.Sel.Name)
		}
	} else {
		t.Error("Multiple elements should create ast.SelectorExpr")
	}

	// Three elements -> nested selector
	builder = NewSimpleTypeBuilder().AddElements("package", "subpackage", "Type")
	expr = builder.Build()

	if selector, ok := expr.(*ast.SelectorExpr); ok {
		if selector.Sel.Name != "Type" {
			t.Errorf("Expected outermost selector name 'Type', got %s", selector.Sel.Name)
		}

		if innerSelector, ok := selector.X.(*ast.SelectorExpr); ok {
			if innerSelector.Sel.Name != "subpackage" {
				t.Errorf("Expected inner selector name 'subpackage', got %s", innerSelector.Sel.Name)
			}
		} else {
			t.Error("Inner expression should be ast.SelectorExpr")
		}
	} else {
		t.Error("Three elements should create ast.SelectorExpr")
	}
}

func TestSimpleTypeBuilder_BuildWithoutElements(t *testing.T) {
	builder := NewSimpleTypeBuilder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Build should panic when no elements are present")
		}
	}()

	builder.Build()
}

func TestSliceOf_Nil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("SliceOf should panic when element is nil")
		}
	}()

	SliceOf(nil)
}

func TestArrayTypeBuilder_Build(t *testing.T) {
	expr := SliceOf(String()).Build()

	if arrayType, ok := expr.(*ast.ArrayType); ok {
		if ident, ok := arrayType.Elt.(*ast.Ident); ok {
			if ident.Name != "string" {
				t.Errorf("Expected element type 'string', got %s", ident.Name)
			}
		} else {
			t.Error("Element type should be ast.Ident")
		}

		if arrayType.Len != nil {
			t.Error("Array type should have nil length for slices")
		}
	} else {
		t.Error("Build should create ast.ArrayType")
	}
}

func TestArrayTypeBuilder_BuildWithoutElement(t *testing.T) {
	builder := &ArrayTypeBuilder{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Build should panic when no element is set")
		}
	}()

	builder.Build()
}

func TestArrayTypeBuilder_NestedArrays(t *testing.T) {
	// [][]string
	expr := SliceOf(SliceOf(String())).Build()

	if outerArrayType, ok := expr.(*ast.ArrayType); ok {
		if innerArrayType, ok := outerArrayType.Elt.(*ast.ArrayType); ok {
			if ident, ok := innerArrayType.Elt.(*ast.Ident); ok {
				if ident.Name != "string" {
					t.Errorf("Expected inner element type 'string', got %s", ident.Name)
				}
			} else {
				t.Error("Inner element type should be ast.Ident")
			}
		} else {
			t.Error("Inner element should be ast.ArrayType")
		}
	} else {
		t.Error("Outer expression should be ast.ArrayType")
	}
}

func TestAliasOf_Build(t *testing.T) {
	spec := AliasOf("StringAlias", String()).Build()

	if spec.Name.Name != "StringAlias" {
		t.Errorf("Expected name 'StringAlias', got %s", spec.Name.Name)
	}

	if ident, ok := spec.Type.(*ast.Ident); ok {
		if ident.Name != "string" {
			t.Errorf("Expected type 'string', got %s", ident.Name)
		}
	} else {
		t.Error("Type should be ast.Ident")
	}
}

func TestAliasOf_BuildWithArrayType(t *testing.T) {
	spec := AliasOf("StringSliceAlias", SliceOf(String())).Build()

	if spec.Name.Name != "StringSliceAlias" {
		t.Errorf("Expected name 'StringSliceAlias', got %s", spec.Name.Name)
	}

	if arrayType, ok := spec.Type.(*ast.ArrayType); ok {
		if ident, ok := arrayType.Elt.(*ast.Ident); ok {
			if ident.Name != "string" {
				t.Errorf("Expected array element type 'string', got %s", ident.Name)
			}
		} else {
			t.Error("Array element should be ast.Ident")
		}
	} else {
		t.Error("Type should be ast.ArrayType")
	}
}

func TestAliasOf_EmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("AliasOf should panic when name is empty")
		}
	}()

	AliasOf("", String())
}

func TestAliasOf_NilType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("AliasOf should panic when type is nil")
		}
	}()

	AliasOf("Name", nil)
}

func TestTypeAliasBuilder_BuildDecl(t *testing.T) {
	decl, ok := AliasOf("MyAlias", String()).BuildDecl().(*ast.GenDecl)
	if !ok {
		t.Fatal("BuildDecl should return *ast.GenDecl")
	}

	if decl.Tok != token.TYPE {
		t.Error("Declaration should have TYPE token")
	}

	if len(decl.Specs) != 1 {
		t.Error("Declaration should have exactly one spec")
	}

	if typeSpec, ok := decl.Specs[0].(*ast.TypeSpec); ok {
		if typeSpec.Name.Name != "MyAlias" {
			t.Errorf("Expected spec name 'MyAlias', got %s", typeSpec.Name.Name)
		}
	} else {
		t.Error("Spec should be ast.TypeSpec")
	}
}

func TestAliasOf_NestedArrayType(t *testing.T) {
	// type StringMatrix [][]string
	spec := AliasOf("StringMatrix", SliceOf(SliceOf(String()))).Build()

	if outerArrayType, ok := spec.Type.(*ast.ArrayType); ok {
		if innerArrayType, ok := outerArrayType.Elt.(*ast.ArrayType); ok {
			if ident, ok := innerArrayType.Elt.(*ast.Ident); ok {
				if ident.Name != "string" {
					t.Errorf("Expected inner array element type 'string', got %s", ident.Name)
				}
			} else {
				t.Error("Inner array element should be ast.Ident")
			}
		} else {
			t.Error("Inner element should be ast.ArrayType")
		}
	} else {
		t.Error("Outer type should be ast.ArrayType")
	}
}

func TestMapOf_Build(t *testing.T) {
	expr := MapOf(String(), Int()).Build()

	mapType, ok := expr.(*ast.MapType)
	if !ok {
		t.Fatal("Build should return *ast.MapType")
	}

	if ident, ok := mapType.Key.(*ast.Ident); !ok || ident.Name != "string" {
		t.Error("Expected key type 'string'")
	}

	if ident, ok := mapType.Value.(*ast.Ident); !ok || ident.Name != "int" {
		t.Error("Expected value type 'int'")
	}
}

func TestMapOf_NilKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MapOf should panic when key is nil")
		}
	}()

	MapOf(nil, Int())
}

func TestMapOf_NilValue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MapOf should panic when value is nil")
		}
	}()

	MapOf(String(), nil)
}

func TestMapTypeBuilder_BuildWithoutKey(t *testing.T) {
	builder := &MapTypeBuilder{value: Int()}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Build should panic when no key is set")
		}
	}()

	builder.Build()
}

func TestMapTypeBuilder_BuildWithoutValue(t *testing.T) {
	builder := &MapTypeBuilder{key: String()}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Build should panic when no value is set")
		}
	}()

	builder.Build()
}
