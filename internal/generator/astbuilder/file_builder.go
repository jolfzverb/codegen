package astbuilder

import (
	"go/ast"
	"go/token"
)

// DeclBuilder builds a top-level Go declaration.
type DeclBuilder interface {
	Build() ast.Decl
}

// TypeSpecBuilder builds an *ast.TypeSpec.
type TypeSpecBuilder interface {
	Build() *ast.TypeSpec
}

// TypeDeclBuilder wraps a TypeSpecBuilder into a type declaration.
type TypeDeclBuilder struct {
	spec *ast.TypeSpec
}

// TypeDecl creates a TypeDeclBuilder from a TypeSpecBuilder.
func TypeDecl(builder TypeSpecBuilder) *TypeDeclBuilder {
	return &TypeDeclBuilder{spec: builder.Build()}
}

// Build creates the ast.Decl.
func (tdb *TypeDeclBuilder) Build() ast.Decl {
	return &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{tdb.spec},
	}
}

// FileBuilder builds an *ast.File with a package declaration, imports, and top-level declarations.
type FileBuilder struct {
	packageName string
	importSpecs []*ast.ImportSpec
	declSpecs   []ast.Spec
	decls       []ast.Decl
}

// NewFileBuilder creates a FileBuilder for the given package name.
func NewFileBuilder(packageName string) *FileBuilder {
	return &FileBuilder{packageName: packageName}
}

// WithImports sets the imports for the file using an ImportsBuilder.
func (fb *FileBuilder) WithImports(ib *ImportsBuilder) *FileBuilder {
	fb.importSpecs, fb.declSpecs = ib.Build()
	return fb
}

// AddDecl appends a top-level declaration to the file.
func (fb *FileBuilder) AddDecl(b DeclBuilder) *FileBuilder {
	fb.decls = append(fb.decls, b.Build())
	return fb
}

// Build assembles and returns the *ast.File.
func (fb *FileBuilder) Build() *ast.File {
	file := &ast.File{
		Name:    ast.NewIdent(fb.packageName),
		Imports: fb.importSpecs,
		Decls:   make([]ast.Decl, 0, len(fb.decls)+1),
	}
	if len(fb.declSpecs) > 0 {
		file.Decls = append(file.Decls, &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: fb.declSpecs,
		})
	}
	for _, decl := range fb.decls {
		file.Decls = append(file.Decls, decl)
	}
	return file
}
