package astbuilder

import (
	"go/ast"
	"go/token"
)

// DeclBuilder builds a top-level Go declaration.
type DeclBuilder interface {
	BuildDecl() ast.Decl
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
	fb.decls = append(fb.decls, b.BuildDecl())
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
