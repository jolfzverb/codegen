package generator

import (
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"strings"

	"github.com/go-faster/errors"
)

type SchemasFile struct {
	file *ast.File
}

type SchemaStruct struct {
	Name   string
	Fields []SchemaField
}

type SchemaField struct {
	Name        string
	Type        string
	TagJson     []string
	TagValidate []string
}

func NewSchemasFile() *SchemasFile {
	return &SchemasFile{
		file: &ast.File{
			Name:    ast.NewIdent("models"),
			Imports: []*ast.ImportSpec{},
			Decls:   []ast.Decl{},
		},
	}
}

func (m *SchemasFile) WriteToOutput(output io.Writer) error {
	const op = "generator.SchemasFile.WriteToOutput"
	err := format.Node(output, token.NewFileSet(), m.file)
	if err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}

func (m *SchemasFile) AddSchema(model SchemaStruct) {
	var fieldList []*ast.Field
	for _, field := range model.Fields {
		jsonTags := strings.Join(field.TagJson, ",")
		validateTags := strings.Join(field.TagValidate, ",")

		var tags string
		if len(field.TagJson) > 0 {
			tags += "json:\"" + jsonTags + "\""
		}
		if len(field.TagValidate) > 0 {
			if len(tags) > 0 {
				tags += " "
			}
			tags += "validate:\"" + validateTags + "\""
		}
		if len(tags) > 0 {
			tags = "`" + tags + "`"
		}

		fieldList = append(fieldList, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(field.Name)},
			Type: &ast.StarExpr{
				Star: token.NoPos,
				X:    ast.NewIdent(field.Type),
			},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: tags,
			},
		})
	}

	m.file.Decls = append(m.file.Decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(model.Name),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fieldList,
					},
				},
			},
		},
	})
}

func (m *SchemasFile) AddTypeAlias(name string, typeName string) {
	m.file.Decls = append(m.file.Decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(name),
				Type: &ast.Ident{
					Name: typeName,
				},
			},
		},
	})
}

func (m *SchemasFile) AddSliceAlias(name string, typeName string) {
	m.file.Decls = append(m.file.Decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(name),
				Type: &ast.ArrayType{
					Elt: ast.NewIdent(typeName),
				},
			},
		},
	})
}
