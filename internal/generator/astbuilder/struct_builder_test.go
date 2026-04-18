package astbuilder

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestNewStructBuilder(t *testing.T) {
	builder := NewStructBuilder()

	if builder == nil {
		t.Fatal("NewStructBuilder returned nil")
	}

	if builder.name != "" {
		t.Errorf("Expected empty name initially, got %s", builder.name)
	}

	if len(builder.fields) != 0 {
		t.Errorf("Expected empty fields initially, got %d fields", len(builder.fields))
	}
}

func TestStruct_Constructor(t *testing.T) {
	sb := Struct("Person", StringField("name"), IntField("age"))

	if sb.name != "Person" {
		t.Errorf("Expected name 'Person', got %s", sb.name)
	}

	if len(sb.fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(sb.fields))
	}
}

func TestStruct_ConstructorNoFields(t *testing.T) {
	sb := Struct("Empty")

	if sb.name != "Empty" {
		t.Errorf("Expected name 'Empty', got %s", sb.name)
	}

	if len(sb.fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(sb.fields))
	}
}

func TestStruct_ConstructorNilField(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Struct should panic when a field is nil")
		}
	}()

	Struct("Bad", StringField("ok"), nil)
}

func TestStructBuilder_WithName(t *testing.T) {
	builder := NewStructBuilder()

	result := builder.WithName("TestStruct")
	if result != builder {
		t.Error("WithName should return the builder for chaining")
	}

	if builder.name != "TestStruct" {
		t.Errorf("Expected name 'TestStruct', got %s", builder.name)
	}
}

func TestStructBuilder_AddFields(t *testing.T) {
	builder := NewStructBuilder()

	result := builder.AddFields(StringField("name"), IntField("age"), BoolField("active"))
	if result != builder {
		t.Error("AddFields should return the builder for chaining")
	}

	if len(builder.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(builder.fields))
	}
}

func TestStructBuilder_AddFieldsNil(t *testing.T) {
	builder := NewStructBuilder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("AddFields should panic when any field builder is nil")
		}
	}()

	builder.AddFields(StringField("name"), nil, IntField("age"))
}

func TestStructBuilder_Build(t *testing.T) {
	typeSpec := Struct("Person", StringField("name"), IntField("age")).Build()

	if typeSpec.Name.Name != "Person" {
		t.Errorf("Expected type name 'Person', got %s", typeSpec.Name.Name)
	}

	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		t.Fatal("Type should be *ast.StructType")
	}

	if len(structType.Fields.List) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(structType.Fields.List))
	}

	field1 := structType.Fields.List[0]
	if len(field1.Names) != 1 || field1.Names[0].Name != "name" {
		t.Error("First field should be named 'name'")
	}
	if ident, ok := field1.Type.(*ast.Ident); !ok || ident.Name != "string" {
		t.Error("First field should be of type 'string'")
	}

	field2 := structType.Fields.List[1]
	if len(field2.Names) != 1 || field2.Names[0].Name != "age" {
		t.Error("Second field should be named 'age'")
	}
	if ident, ok := field2.Type.(*ast.Ident); !ok || ident.Name != "int" {
		t.Error("Second field should be of type 'int'")
	}
}

func TestStructBuilder_BuildWithoutName(t *testing.T) {
	builder := NewStructBuilder().AddFields(StringField("name"))

	defer func() {
		if r := recover(); r == nil {
			t.Error("Build should panic when struct has no name")
		}
	}()

	builder.Build()
}

func TestStructBuilder_TypeDecl(t *testing.T) {
	decl, ok := TypeDecl(Struct("Person", StringField("name"))).Build().(*ast.GenDecl)
	if !ok {
		t.Fatal("TypeDecl.Build should return *ast.GenDecl")
	}

	if decl.Tok != token.TYPE {
		t.Error("Declaration should have token.TYPE")
	}

	if len(decl.Specs) != 1 {
		t.Errorf("Expected 1 spec, got %d", len(decl.Specs))
	}

	typeSpec, ok := decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		t.Fatal("Spec should be *ast.TypeSpec")
	}

	if typeSpec.Name.Name != "Person" {
		t.Errorf("Expected type name 'Person', got %s", typeSpec.Name.Name)
	}
}

func TestStructBuilder_MethodChaining(t *testing.T) {
	builder := Struct("Person").
		AddFields(StringField("name"), IntField("age")).
		AddFields(BoolField("active"))

	if builder.name != "Person" {
		t.Errorf("Expected name 'Person', got %s", builder.name)
	}

	if len(builder.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(builder.fields))
	}
}

func TestStructBuilder_ComplexExample(t *testing.T) {
	typeSpec := Struct("CreateRequest",
		StringField("name"),
		StringField("description"),
		IntField("priority"),
		BoolField("active"),
		ContextField("ctx"),
		Field("metadata", SimpleType("apimodels", "Metadata")),
		NewFieldBuilder().WithName("tags").WithType(String()).AddJSONTags("tags", "omitempty"),
	).Build()

	if typeSpec.Name.Name != "CreateRequest" {
		t.Errorf("Expected type name 'CreateRequest', got %s", typeSpec.Name.Name)
	}

	structType := typeSpec.Type.(*ast.StructType)
	if len(structType.Fields.List) != 7 {
		t.Errorf("Expected 7 fields, got %d", len(structType.Fields.List))
	}

	fields := structType.Fields.List
	if fields[0].Names[0].Name != "name" {
		t.Error("First field should be named 'name'")
	}
	if fields[5].Names[0].Name != "metadata" {
		t.Error("Sixth field should be named 'metadata'")
	}
	if fields[6].Names[0].Name != "tags" {
		t.Error("Seventh field should be named 'tags'")
	}
}
