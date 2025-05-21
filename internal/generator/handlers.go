package generator

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"

	"github.com/go-faster/errors"
)

type HandlersFile struct {
	packageName    *ast.Ident
	packageImports []*ast.ImportSpec
	importDecl     *ast.GenDecl
	interfaceDecls []*ast.GenDecl

	handlerDecl            *ast.GenDecl
	handlerDeclQAFieldList *ast.FieldList // quick access to handler struct field list

	handlerConstructorDecl                       *ast.FuncDecl
	handlerConstructorDeclQAArgs                 *ast.FieldList    // quick access to handler constructor args
	handlerConstructorDeclQAConstructorComposite *ast.CompositeLit // quick access to handler struct initializer

	addRoutesDecl *ast.FuncDecl
	handleDecls   []*ast.FuncDecl
}

func (h *HandlersFile) InitImports(modelsImportPath string) {
	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("%q", "github.com/go-playground/validator/v10"),
		},
	}
	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importSpec},
	}
	h.importDecl = importDecl
	h.packageImports = append(h.packageImports, importSpec)

	h.AddImport(modelsImportPath)
	h.AddImport("context")
	h.AddImport("github.com/go-chi/chi/v5")
}

func (h *HandlersFile) InitHandlerStruct() {
	fieldList := &ast.FieldList{
		List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent("validator")},
			Type: &ast.StarExpr{
				Star: token.NoPos,
				X: &ast.SelectorExpr{
					X:   ast.NewIdent("validator"),
					Sel: ast.NewIdent("Validate"),
				},
			},
		}},
	}
	handlerDecl := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent("Handler"),
				Type: &ast.StructType{
					Fields: fieldList,
				},
			},
		},
	}
	h.handlerDecl = handlerDecl
	h.handlerDeclQAFieldList = fieldList
}

func (h *HandlersFile) InitHandlerConstructor() {
	initializerComposite := &ast.CompositeLit{
		Type: &ast.Ident{
			Name: "Handler",
		},
		Elts: []ast.Expr{
			&ast.KeyValueExpr{
				Key: ast.NewIdent("validator"),
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("validator"),
						Sel: ast.NewIdent("New"),
					},
					Args: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("validator"),
								Sel: ast.NewIdent("WithRequiredStructEnabled"),
							},
						},
					},
				},
			},
		},
	}

	newHandlerDecl := &ast.FuncDecl{
		Name: ast.NewIdent("NewHandler"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{{
					Type: &ast.StarExpr{
						Star: token.NoPos,
						X: &ast.Ident{
							Name: "Handler",
						},
					},
				}},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X:  initializerComposite,
						},
					},
				},
			},
		},
	}

	h.handlerConstructorDeclQAArgs = newHandlerDecl.Type.Params
	h.handlerConstructorDeclQAConstructorComposite = initializerComposite
	h.handlerConstructorDecl = newHandlerDecl
}

func (h *HandlersFile) InitRoutesFunc() {
	addRoutesDecl := &ast.FuncDecl{
		Name: ast.NewIdent("AddRoutes"),
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("h")},
				Type: &ast.StarExpr{
					Star: token.NoPos,
					X:    ast.NewIdent("Handler"),
				},
			}},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("router")},
					Type: &ast.StarExpr{
						Star: token.NoPos,
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("chi"),
							Sel: ast.NewIdent("Router"),
						},
					},
				}},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
	h.addRoutesDecl = addRoutesDecl
}

func (h *HandlersFile) InitFields(packageName string, modelsImportPath string) {
	// package
	h.packageName = ast.NewIdent(packageName)

	// imports
	h.InitImports(modelsImportPath)

	// list of interfaces

	// handler
	h.InitHandlerStruct()

	// handler constructor
	h.InitHandlerConstructor()

	// add routes func
	h.InitRoutesFunc()

	// parse functions
	// handle functions
	// handle by content type
}

func NewHandlersFile(packageName string, modelsImportPath string) *HandlersFile {
	h := &HandlersFile{}
	h.InitFields(packageName, modelsImportPath)

	return h
}

func (h *HandlersFile) WriteToOutput(output io.Writer) error {
	const op = "generator.HandlersFile.WriteToOutput"
	file := h.GenerateFile()
	err := format.Node(output, token.NewFileSet(), file)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (h *HandlersFile) AddInterface(name string, methodName string, requestName string, responseName string) {
	var methodParams []*ast.Field
	methodParams = append(methodParams, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("ctx")},
		Type:  ast.NewIdent("context.Context"),
	})
	methodParams = append(methodParams, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("r")},
		Type: &ast.StarExpr{
			Star: token.NoPos,
			X: &ast.SelectorExpr{
				X:   ast.NewIdent("models"),
				Sel: ast.NewIdent(requestName),
			},
		},
	})
	var methodResults []*ast.Field
	methodResults = append(methodResults, &ast.Field{
		Type: &ast.StarExpr{
			Star: token.NoPos,
			X: &ast.SelectorExpr{
				X:   ast.NewIdent("models"),
				Sel: ast.NewIdent(responseName),
			},
		},
	})
	methodResults = append(methodResults, &ast.Field{
		Type: ast.NewIdent("error"),
	})
	h.interfaceDecls = append(h.interfaceDecls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(name),
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{
						List: []*ast.Field{{
							Names: []*ast.Ident{ast.NewIdent(methodName)},
							Type: &ast.FuncType{
								Params: &ast.FieldList{
									List: methodParams,
								},
								Results: &ast.FieldList{
									List: methodResults,
								},
							},
						}},
					},
				},
			},
		},
	})
}

func (h *HandlersFile) AddDependencyToHandler(baseName string) {
	fieldName := GoIdentLowercase(baseName)

	h.handlerDeclQAFieldList.List = append(h.handlerDeclQAFieldList.List, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(fieldName)},
		Type:  ast.NewIdent(baseName + "Handler"),
	})

	h.handlerConstructorDeclQAArgs.List = append(h.handlerConstructorDeclQAArgs.List, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(fieldName)},
		Type:  ast.NewIdent(baseName + "Handler"),
	})

	h.handlerConstructorDeclQAConstructorComposite.Elts = append(
		h.handlerConstructorDeclQAConstructorComposite.Elts, &ast.KeyValueExpr{
			Key:   ast.NewIdent(fieldName),
			Value: ast.NewIdent(fieldName),
		},
	)
}

func (h *HandlersFile) AddImport(path string) {
	for _, imp := range h.packageImports {
		if imp.Path.Value == fmt.Sprintf("%q", path) {
			return
		}
	}
	imp := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("%q", path),
		},
	}
	h.packageImports = append(h.packageImports, imp)
	h.importDecl.Specs = append(h.importDecl.Specs, imp)
}

func (h *HandlersFile) GenerateFile() *ast.File {
	file := &ast.File{
		Name:    h.packageName,
		Decls:   []ast.Decl{},
		Imports: h.packageImports,
	}

	file.Decls = append(file.Decls, h.importDecl)
	for _, d := range h.interfaceDecls {
		file.Decls = append(file.Decls, d)
	}

	file.Decls = append(file.Decls, h.handlerDecl)
	file.Decls = append(file.Decls, h.handlerConstructorDecl)
	file.Decls = append(file.Decls, h.addRoutesDecl)
	for _, d := range h.handleDecls {
		file.Decls = append(file.Decls, d)
	}

	return file
}

func (h *HandlersFile) AddRouteToRouter(baseName string, method string, pathName string) {
	h.addRoutesDecl.Body.List = append(h.addRoutesDecl.Body.List, &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("router"),
				Sel: ast.NewIdent(method),
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("%q", pathName),
				},
				&ast.SelectorExpr{
					X:   ast.NewIdent("h"),
					Sel: ast.NewIdent("handle" + baseName),
				},
			},
		},
	})
}
