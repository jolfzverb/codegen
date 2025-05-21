package generator_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/jolfzverb/codegen/internal/generator"
	"github.com/stretchr/testify/assert"
)

func TestGenerateValidInput(t *testing.T) {
	mockInput := `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ExampleModel:
      type: object
      properties:
        field_two:
          type: integer
        field_one:
          type: string
          minLength: 3
          maxLength: 10
        field_three:
          $ref: '#/components/schemas/StringModel'
        field4:
          type: object
          properties:
            field1:
              type: string
        field5:
          $ref: '#/components/schemas/ObjectModel'
        field6:
          type: number
          minimum: 1.5
          maximum: 10.5
      required:
        - field_one
        - field6
    StringModel:
      type: string
    IntModel:
      type: integer
    ObjectModel:
      type: object
      properties:
        field1:
          type: string
`

	expectedOutput := `package models

type ExampleModelField4 struct {
	Field1 *string ` + "`json:\"field1,omitempty\" validate:\"omitempty\"`" + `
}
type ExampleModel struct {
	Field4     *ExampleModelField4 ` + "`json:\"field4,omitempty\" validate:\"omitempty\"`" + `
	Field5     *ObjectModel        ` + "`json:\"field5,omitempty\" validate:\"omitempty\"`" + `
	Field6     *float64            ` + "`json:\"field6\" validate:\"required,min=1.5,max=10.5\"`" + `
	FieldOne   *string             ` + "`json:\"field_one\" validate:\"required,min=3,max=10\"`" + `
	FieldThree *StringModel        ` + "`json:\"field_three,omitempty\" validate:\"omitempty\"`" + `
	FieldTwo   *int                ` + "`json:\"field_two,omitempty\" validate:\"omitempty\"`" + `
}
type IntModel int
type ObjectModel struct {
	Field1 *string ` + "`json:\"field1,omitempty\" validate:\"omitempty\"`" + `
}
type StringModel string
`

	input := strings.NewReader(mockInput)
	outputModels := &bytes.Buffer{}
	outputHandlers := &bytes.Buffer{}

	err := generator.GenerateToIO(context.Background(), input, outputModels, outputHandlers, "imports", "packagename")

	assert.NoError(t, err)

	assert.Equal(t, expectedOutput, outputModels.String())
}

func TestGeneratorFeatures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "TestString",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    StringModel:
      type: string
`,
			expected: `package models

type StringModel string
`,
		},
		{
			name: "TestInt",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    IntModel:
      type: integer
`,
			expected: `package models

type IntModel int
`,
		},
		{
			name: "TestNumber",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    NumberModel:
      type: number
`,
			expected: `package models

type NumberModel float64
`,
		},
		{
			name: "TestBool",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    BoolModel:
      type: boolean
`,
			expected: `package models

type BoolModel bool
`,
		},
		{
			name: "TestSimpleObject",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ObjectModel:
      type: object
      properties:
        string_field:
          type: string
`,
			expected: `package models

type ObjectModel struct {
	StringField *string ` + "`json:\"string_field,omitempty\" validate:\"omitempty\"`" + `
}
`,
		},
		{
			name: "TestEmptyObject",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ObjectModel:
      type: object
`,
			expected: `package models

type ObjectModel struct {
}
`,
		},
		{
			name: "TestStringValidatorsObject",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ObjectModel:
      type: object
      properties:
        string_field:
          type: string
          minLength: 3
          maxLength: 10
`,
			expected: `package models

type ObjectModel struct {
	StringField *string ` + "`json:\"string_field,omitempty\" validate:\"omitempty,min=3,max=10\"`" + `
}
`,
		},
		{
			name: "TestStringReferenceObject",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    StringModel:
      type: string
      minLength: 3
      maxLength: 10
    ObjectModel:
      type: object
      properties:
        string_field:
          $ref: '#/components/schemas/StringModel'
`,
			expected: `package models

type ObjectModel struct {
	StringField *StringModel ` + "`json:\"string_field,omitempty\" validate:\"omitempty,min=3,max=10\"`" + `
}
type StringModel string
`,
		},
		{
			name: "TestFieldTypesInObject",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ObjectModel:
      type: object
      properties:
        string_field:
          type: string
        int_field:
          type: integer
        number_field:
          type: number
        bool_field:
          type: boolean
`,
			expected: `package models

type ObjectModel struct {
	BoolField   *bool    ` + "`json:\"bool_field,omitempty\" validate:\"omitempty\"`" + `
	IntField    *int     ` + "`json:\"int_field,omitempty\" validate:\"omitempty\"`" + `
	NumberField *float64 ` + "`json:\"number_field,omitempty\" validate:\"omitempty\"`" + `
	StringField *string  ` + "`json:\"string_field,omitempty\" validate:\"omitempty\"`" + `
}
`,
		},
		{
			name: "TestObjectRef",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ObjectModel:
      type: object
      properties:
        object_field:
          $ref: '#/components/schemas/Object2Model'
    Object2Model:
      type: object
      properties:
        string_field:
          type: string
`,
			expected: `package models

type Object2Model struct {
	StringField *string ` + "`json:\"string_field,omitempty\" validate:\"omitempty\"`" + `
}
type ObjectModel struct {
	ObjectField *Object2Model ` + "`json:\"object_field,omitempty\" validate:\"omitempty\"`" + `
}
`,
		},
		{
			name: "TestNestedObject",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ObjectModel:
      type: object
      properties:
        object_field:
          type: object
          properties:
            string_field:
              type: string
`,
			expected: `package models

type ObjectModelObjectField struct {
	StringField *string ` + "`json:\"string_field,omitempty\" validate:\"omitempty\"`" + `
}
type ObjectModel struct {
	ObjectField *ObjectModelObjectField ` + "`json:\"object_field,omitempty\" validate:\"omitempty\"`" + `
}
`,
		},
		{
			name: "TestArrayStringRefSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        $ref: '#/components/schemas/StringModel'
    StringModel:
      type: string
`,
			expected: `package models

type ArrayModel []StringModel
type StringModel string
`,
		},
		{
			name: "TestArrayObjectRefSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        $ref: '#/components/schemas/ObjectModel'
    ObjectModel:
      type: object
      properties:
        string_field:
          type: string
`,
			expected: `package models

type ArrayModel []ObjectModel
type ObjectModel struct {
	StringField *string ` + "`json:\"string_field,omitempty\" validate:\"omitempty\"`" + `
}
`,
		},
		{
			name: "TestArrayStringSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        type: string
`,
			expected: `package models

type ArrayModel []string
`,
		},
		{
			name: "TestArrayBoolSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        type: boolean
`,
			expected: `package models

type ArrayModel []bool
`,
		},
		{
			name: "TestArrayIntSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        type: integer
`,
			expected: `package models

type ArrayModel []int
`,
		},
		{
			name: "TestArrayNumberSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        type: number
`,
			expected: `package models

type ArrayModel []float64
`,
		},
		{
			name: "TestArrayObjectSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        type: object
        properties:
          string_field:
            type: string
`,
			expected: `package models

type ArrayModelItem struct {
	StringField *string ` + "`json:\"string_field,omitempty\" validate:\"omitempty\"`" + `
}
type ArrayModel []ArrayModelItem
`,
		},
		{
			name: "TestArrayNestedObjectSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        type: object
        properties:
          object_field:
            type: object
            properties:
              string_field:
                type: string
`,
			expected: `package models

type ArrayModelItemObjectField struct {
	StringField *string ` + "`json:\"string_field,omitempty\" validate:\"omitempty\"`" + `
}
type ArrayModelItem struct {
	ObjectField *ArrayModelItemObjectField ` + "`json:\"object_field,omitempty\" validate:\"omitempty\"`" + `
}
type ArrayModel []ArrayModelItem
`,
		},
		{
			name: "TestArrayNestedArrayRefSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        $ref: '#/components/schemas/Array2Model'
    Array2Model:
      type: array
      items:
        type: string
`,
			expected: `package models

type Array2Model []string
type ArrayModel []Array2Model
`,
		},
		{
			name: "TestArrayNestedArraySchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        type: array
        items:
          type: string
`,
			expected: `package models

type ArrayModelItem []string
type ArrayModel []ArrayModelItem
`,
		},
		{
			name: "TestObjectRefArraySchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ArrayModel:
      type: array
      items:
        type: string
    ObjectModel:
      type: object
      properties:
        array_field:
          $ref: '#/components/schemas/ArrayModel'
`,
			expected: `package models

type ArrayModel []string
type ObjectModel struct {
	ArrayField *ArrayModel ` + "`json:\"array_field,omitempty\" validate:\"omitempty\"`" + `
}
`,
		},
		{
			name: "TestObjectArraySchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ObjectModel:
      type: object
      properties:
        array_field:
          type: array
          items:
            type: string
`,
			expected: `package models

type ObjectModelArrayField []string
type ObjectModel struct {
	ArrayField *ObjectModelArrayField ` + "`json:\"array_field,omitempty\" validate:\"omitempty\"`" + `
}
`,
		},
		{
			name: "TestObjectArrayValidatorsSchema",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths: {}
components:
  schemas:
    ObjectModel:
      type: object
      properties:
        array_field:
          type: array
          minItems: 1
          maxItems: 10
          uniqueItems: true
          items:
            type: string
            minLength: 3
            maxLength: 10
`,
			expected: `package models

type ObjectModelArrayField []string
type ObjectModel struct {
	ArrayField *ObjectModelArrayField ` + "`json:\"array_field,omitempty\" validate:\"omitempty,min=1,max=10,unique,dive,min=3,max=10\"`" + `
}
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := strings.NewReader(tc.input)
			outputModels := &bytes.Buffer{}
			outputHandlers := &bytes.Buffer{}

			err := generator.GenerateToIO(context.Background(), input, outputModels, outputHandlers, "imports", "packagename")

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, outputModels.String())
		})
	}
}

func TestGeneratePaths(t *testing.T) {
	for _, tc := range []struct {
		name             string
		input            string
		expectedModels   string
		expectedHandlers string
	}{
		{
			name: "TestString",
			input: `openapi: 3.0.0
info:
  title: API
  version: 1.0.0
paths:
  /example:
    get:
      summary: Example
      responses:
        '200':
          description: OK
  /example2:
    get:
      summary: Example
      responses:
        '200':
          description: OK
`,
			expectedModels: `package models
`,
			expectedHandlers: `package packagename

import (
	"github.com/go-playground/validator/v10"
	"imports/models"
	"context"
	"github.com/go-chi/chi/v5"
)

type GetExample2JsonHandler interface {
	HandleGetExample2Json(ctx context.Context, r *models.GetExample2JsonRequest) (*models.GetExample2JsonResponse, error)
}
type GetExampleJsonHandler interface {
	HandleGetExampleJson(ctx context.Context, r *models.GetExampleJsonRequest) (*models.GetExampleJsonResponse, error)
}
type Handler struct {
	validator       *validator.Validate
	getExample2Json GetExample2JsonHandler
	getExampleJson  GetExampleJsonHandler
}

func NewHandler(getExample2Json GetExample2JsonHandler, getExampleJson GetExampleJsonHandler) *Handler {
	return &Handler{validator: validator.New(validator.WithRequiredStructEnabled()), getExample2Json: getExample2Json, getExampleJson: getExampleJson}
}
func (h *Handler) AddRoutes(router chi.Router) {
	router.Get("/example2", h.handleGetExample2Json)
	router.Get("/example", h.handleGetExampleJson)
}
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := strings.NewReader(tc.input)
			outputModels := &bytes.Buffer{}
			outputHandlers := &bytes.Buffer{}

			err := generator.GenerateToIO(context.Background(), input, outputModels, outputHandlers, "imports", "packagename")

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedModels, outputModels.String())
			assert.Equal(t, tc.expectedHandlers, outputHandlers.String())
		})
	}
}
