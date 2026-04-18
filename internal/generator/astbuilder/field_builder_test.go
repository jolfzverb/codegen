package astbuilder

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestNewFieldBuilder(t *testing.T) {
	builder := NewFieldBuilder()

	if builder == nil {
		t.Fatal("NewFieldBuilder returned nil")
	}

	if builder.name != "" {
		t.Errorf("Expected empty name initially, got %s", builder.name)
	}

	if builder.typeBuilder == nil {
		t.Fatal("typeBuilder should not be nil")
	}

	if len(builder.jsonTags) != 0 {
		t.Errorf("Expected empty jsonTags initially, got %v", builder.jsonTags)
	}

	if len(builder.validateTags) != 0 {
		t.Errorf("Expected empty validateTags initially, got %v", builder.validateTags)
	}
}

func TestFieldBuilder_WithName(t *testing.T) {
	builder := NewFieldBuilder()

	result := builder.WithName("fieldName")
	if result != builder {
		t.Error("WithName should return the builder for chaining")
	}

	if builder.name != "fieldName" {
		t.Errorf("Expected name 'fieldName', got %s", builder.name)
	}
}

func TestFieldBuilder_WithType(t *testing.T) {
	builder := NewFieldBuilder()
	typeBuilder := String()

	result := builder.WithType(typeBuilder)
	if result != builder {
		t.Error("WithType should return the builder for chaining")
	}

	if builder.typeBuilder == nil {
		t.Fatal("typeBuilder should not be nil")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("WithType should panic when type builder is nil")
		}
	}()

	builder.WithType(nil)
}

func TestFieldBuilder_AddJSONTags(t *testing.T) {
	builder := NewFieldBuilder()

	result := builder.AddJSONTags("name")
	if result != builder {
		t.Error("AddJSONTags should return the builder for chaining")
	}

	if len(builder.jsonTags) != 1 || builder.jsonTags[0] != "name" {
		t.Errorf("Expected jsonTags ['name'], got %v", builder.jsonTags)
	}

	builder.AddJSONTags("omitempty")
	if len(builder.jsonTags) != 2 || builder.jsonTags[1] != "omitempty" {
		t.Errorf("Expected jsonTags ['name', 'omitempty'], got %v", builder.jsonTags)
	}

	// Empty strings should be ignored
	builder.AddJSONTags("")
	if len(builder.jsonTags) != 2 {
		t.Errorf("Expected 2 jsonTags after adding empty string, got %d", len(builder.jsonTags))
	}
}

func TestFieldBuilder_AddValidateTags(t *testing.T) {
	builder := NewFieldBuilder()

	result := builder.AddValidateTags("required")
	if result != builder {
		t.Error("AddValidateTags should return the builder for chaining")
	}

	if len(builder.validateTags) != 1 || builder.validateTags[0] != "required" {
		t.Errorf("Expected validateTags ['required'], got %v", builder.validateTags)
	}

	builder.AddValidateTags("min=1")
	if len(builder.validateTags) != 2 || builder.validateTags[1] != "min=1" {
		t.Errorf("Expected validateTags ['required', 'min=1'], got %v", builder.validateTags)
	}
}

func TestFieldBuilder_Build(t *testing.T) {
	builder := NewFieldBuilder().
		WithName("fieldName").
		WithType(String()).
		AddJSONTags("name")

	field := builder.Build()

	if field == nil {
		t.Fatal("Build returned nil")
	}

	if len(field.Names) != 1 {
		t.Errorf("Expected 1 name, got %d", len(field.Names))
	}

	if field.Names[0].Name != "fieldName" {
		t.Errorf("Expected name 'fieldName', got %s", field.Names[0].Name)
	}

	if ident, ok := field.Type.(*ast.Ident); ok {
		if ident.Name != "string" {
			t.Errorf("Expected type 'string', got %s", ident.Name)
		}
	} else {
		t.Error("Type should be ast.Ident")
	}

	if field.Tag == nil {
		t.Fatal("Tag should not be nil")
	}

	expectedTag := "`json:\"name\"`"
	if field.Tag.Value != expectedTag {
		t.Errorf("Expected tag '%s', got %s", expectedTag, field.Tag.Value)
	}

	if field.Tag.Kind != token.STRING {
		t.Error("Tag should be a string literal")
	}
}

func TestFieldBuilder_BuildWithoutName(t *testing.T) {
	builder := NewFieldBuilder().WithType(String())
	field := builder.Build()

	if len(field.Names) != 0 {
		t.Errorf("Expected 0 names for unnamed field, got %d", len(field.Names))
	}
}

func TestFieldBuilder_BuildWithoutTag(t *testing.T) {
	builder := NewFieldBuilder().WithName("field").WithType(String())
	field := builder.Build()

	if field.Tag != nil {
		t.Error("Tag should be nil when not set")
	}
}

func TestFieldBuilder_BuildWithMultipleTags(t *testing.T) {
	builder := NewFieldBuilder().
		WithName("field").
		WithType(String()).
		AddJSONTags("name", "omitempty").
		AddValidateTags("required", "min=1")

	field := builder.Build()

	expectedTag := "`json:\"name,omitempty\" validate:\"required,min=1\"`"
	if field.Tag.Value != expectedTag {
		t.Errorf("Expected tag '%s', got %s", expectedTag, field.Tag.Value)
	}
}

func TestFieldBuilder_BuildWithOnlyValidateTags(t *testing.T) {
	builder := NewFieldBuilder().
		WithName("field").
		WithType(String()).
		AddValidateTags("required")

	field := builder.Build()

	expectedTag := "`validate:\"required\"`"
	if field.Tag.Value != expectedTag {
		t.Errorf("Expected tag '%s', got %s", expectedTag, field.Tag.Value)
	}
}

func TestField_Constructor(t *testing.T) {
	// Simple ident type
	field := Field("id", I("UserID")).Build()

	if len(field.Names) != 1 || field.Names[0].Name != "id" {
		t.Error("Field should set the name correctly")
	}

	if ident, ok := field.Type.(*ast.Ident); !ok || ident.Name != "UserID" {
		t.Error("Field should set type to UserID")
	}

	// Selector type
	field = Field("req", SimpleType("models", "Request")).Build()

	if len(field.Names) != 1 || field.Names[0].Name != "req" {
		t.Error("Field should set the name correctly")
	}

	if selector, ok := field.Type.(*ast.SelectorExpr); !ok || selector.Sel.Name != "Request" {
		t.Error("Field should set type to models.Request")
	}
}

func TestFieldBuilder_HelperMethods(t *testing.T) {
	// StringField
	field := StringField("name").Build()
	if ident, ok := field.Type.(*ast.Ident); !ok || ident.Name != "string" {
		t.Error("StringField should set type to string")
	}

	// IntField
	field = IntField("count").Build()
	if ident, ok := field.Type.(*ast.Ident); !ok || ident.Name != "int" {
		t.Error("IntField should set type to int")
	}

	// BoolField
	field = BoolField("active").Build()
	if ident, ok := field.Type.(*ast.Ident); !ok || ident.Name != "bool" {
		t.Error("BoolField should set type to bool")
	}

	// ErrorField
	field = ErrorField().Build()
	if len(field.Names) != 0 {
		t.Error("ErrorField should not have a name")
	}
	if ident, ok := field.Type.(*ast.Ident); !ok || ident.Name != "error" {
		t.Error("ErrorField should set type to error")
	}

	// ContextField
	field = ContextField("ctx").Build()
	if selector, ok := field.Type.(*ast.SelectorExpr); !ok || selector.Sel.Name != "Context" {
		t.Error("ContextField should set type to context.Context")
	}
}

func TestFieldBuilder_WithArrayType(t *testing.T) {
	builder := NewFieldBuilder().
		WithName("tags").
		WithType(SliceOf(String()))

	field := builder.Build()

	if field.Tag != nil {
		t.Error("Tag should be nil when not set")
	}

	if arrayType, ok := field.Type.(*ast.ArrayType); ok {
		if ident, ok := arrayType.Elt.(*ast.Ident); ok {
			if ident.Name != "string" {
				t.Errorf("Expected array element type 'string', got %s", ident.Name)
			}
		} else {
			t.Error("Array element should be ast.Ident")
		}
	} else {
		t.Error("Field type should be ast.ArrayType")
	}
}

func TestFieldBuilder_WithNestedArrayType(t *testing.T) {
	nestedArray := SliceOf(SliceOf(String()))
	builder := NewFieldBuilder().
		WithName("matrix").
		WithType(nestedArray)

	field := builder.Build()

	if outerArrayType, ok := field.Type.(*ast.ArrayType); ok {
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
		t.Error("Outer field type should be ast.ArrayType")
	}
}

func TestFieldBuilder_WithArrayTypeAndTags(t *testing.T) {
	builder := NewFieldBuilder().
		WithName("items").
		WithType(SliceOf(String())).
		AddJSONTags("items", "omitempty").
		AddValidateTags("required", "min=1")

	field := builder.Build()

	expectedTag := "`json:\"items,omitempty\" validate:\"required,min=1\"`"
	if field.Tag.Value != expectedTag {
		t.Errorf("Expected tag '%s', got %s", expectedTag, field.Tag.Value)
	}

	if arrayType, ok := field.Type.(*ast.ArrayType); ok {
		if ident, ok := arrayType.Elt.(*ast.Ident); ok {
			if ident.Name != "string" {
				t.Errorf("Expected array element type 'string', got %s", ident.Name)
			}
		}
	}
}
