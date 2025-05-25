package generator

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"

	"github.com/getkin/kin-openapi/openapi3"
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

	addRoutesDecl        *ast.FuncDecl
	handleDeclQASwitches map[string]*ast.BlockStmt
	restDecls            []*ast.FuncDecl
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
	h.AddImport("net/http")
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
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("chi"),
						Sel: ast.NewIdent("Router"),
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
	h.FinalizeHandlerSwitches()
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
	for _, d := range h.restDecls {
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

func (h *HandlersFile) GetHandler(baseName string) *ast.BlockStmt {
	if h.handleDeclQASwitches == nil {
		return nil
	}
	if blockStmt, ok := h.handleDeclQASwitches[baseName]; ok {
		return blockStmt
	}

	return nil
}

func (h *HandlersFile) CreateHandler(baseName string) {
	switchBody := &ast.BlockStmt{
		List: []ast.Stmt{},
	}

	handleDecl := &ast.FuncDecl{
		Name: ast.NewIdent("handle" + baseName),
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("h")},
				Type: &ast.StarExpr{
					X: ast.NewIdent("Handler"),
				},
			}},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("w")},
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("http"),
						Sel: ast.NewIdent("ResponseWriter"),
					},
				}, {
					Names: []*ast.Ident{ast.NewIdent("r")},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("http"),
							Sel: ast.NewIdent("Request"),
						},
					},
				}},
			},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.SwitchStmt{
					Tag: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("r.Header"),
							Sel: ast.NewIdent("Get"),
						},
						Args: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.STRING,
								Value: `"Content-Type"`,
							},
						},
					},
					Body: switchBody,
				},
			},
		},
	}

	h.restDecls = append(h.restDecls, handleDecl)
	if h.handleDeclQASwitches == nil {
		h.handleDeclQASwitches = make(map[string]*ast.BlockStmt)
	}
	h.handleDeclQASwitches[baseName] = switchBody
}

func (h *HandlersFile) FinalizeHandlerSwitches() {
	if h.handleDeclQASwitches == nil {
		return
	}
	for _, blockStmt := range h.handleDeclQASwitches {
		blockStmt.List = append(blockStmt.List, &ast.CaseClause{
			List: nil,
			Body: []ast.Stmt{
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("http"),
							Sel: ast.NewIdent("Error"),
						},
						Args: []ast.Expr{
							ast.NewIdent("w"),
							&ast.BasicLit{
								Kind:  token.STRING,
								Value: `"Unsupported Content-Type"`,
							},
							&ast.SelectorExpr{
								X:   ast.NewIdent("http"),
								Sel: ast.NewIdent("StatusUnsupportedMediaType"),
							},
						},
					},
				},
				&ast.ReturnStmt{},
			},
		})
	}
}

func (h *HandlersFile) AddContentTypeHandler(baseName string, rawContentType string, handlerSuffix string) {
	if h.handleDeclQASwitches == nil {
		return
	}
	if blockStmt, ok := h.handleDeclQASwitches[baseName]; ok {
		stmts := []ast.Stmt{
			&ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("h"),
						Sel: ast.NewIdent("handle" + baseName + handlerSuffix),
					},
					Args: []ast.Expr{
						ast.NewIdent("w"),
						ast.NewIdent("r"),
					},
				},
			},
			&ast.ReturnStmt{},
		}

		blockStmt.List = append(blockStmt.List, &ast.CaseClause{
			List: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("%q", rawContentType),
				},
			},
			Body: stmts,
		},
		)

		if rawContentType == applicationJSONCT {
			blockStmt.List = append(blockStmt.List, &ast.CaseClause{
				List: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: `""`,
					},
				},
				Body: stmts,
			})
		}
	}
}

func (h *HandlersFile) AddHandleOperationMethod(baseName string) {
	h.restDecls = append(h.restDecls, &ast.FuncDecl{
		Name: ast.NewIdent("handle" + baseName),
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("h")},
				Type: &ast.StarExpr{
					X: ast.NewIdent("Handler"),
				},
			}},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("w")},
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("http"),
						Sel: ast.NewIdent("ResponseWriter"),
					},
				}, {
					Names: []*ast.Ident{ast.NewIdent("r")},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("http"),
							Sel: ast.NewIdent("Request"),
						},
					},
				}},
			},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("request"),
						ast.NewIdent("err"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("h"),
								Sel: ast.NewIdent("parse" + baseName + "Request"),
							},
							Args: []ast.Expr{
								ast.NewIdent("r"),
							},
						},
					},
				},
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent("err"),
						Op: token.NEQ,
						Y:  ast.NewIdent("nil"),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   ast.NewIdent("http"),
										Sel: ast.NewIdent("Error"),
									},
									Args: []ast.Expr{
										ast.NewIdent("w"),
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: `"InternalServerError"`,
										},
										&ast.SelectorExpr{
											X:   ast.NewIdent("http"),
											Sel: ast.NewIdent("StatusInternalServerError"),
										},
									},
								},
							},
							&ast.ReturnStmt{},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("ctx"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("r"),
								Sel: ast.NewIdent("Context"),
							},
							Args: []ast.Expr{},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("response"),
						ast.NewIdent("err"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("h"),
									Sel: ast.NewIdent(GoIdentLowercase(baseName)),
								},
								Sel: ast.NewIdent("Handle" + baseName),
							},
							Args: []ast.Expr{
								ast.NewIdent("ctx"),
								ast.NewIdent("request"),
							},
						},
					},
				},
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X: &ast.BinaryExpr{
							X:  ast.NewIdent("err"),
							Op: token.NEQ,
							Y:  ast.NewIdent("nil"),
						},
						Op: token.OR,
						Y: &ast.BinaryExpr{
							X:  ast.NewIdent("response"),
							Op: token.EQL,
							Y:  ast.NewIdent("nil"),
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   ast.NewIdent("http"),
										Sel: ast.NewIdent("Error"),
									},
									Args: []ast.Expr{
										ast.NewIdent("w"),
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: `"InternalServerError"`,
										},
										&ast.SelectorExpr{
											X:   ast.NewIdent("http"),
											Sel: ast.NewIdent("StatusInternalServerError"),
										},
									},
								},
							},
							&ast.ReturnStmt{},
						},
					},
				},

				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("h"),
							Sel: ast.NewIdent("write" + baseName + "Response"),
						},
						Args: []ast.Expr{
							ast.NewIdent("w"),
							ast.NewIdent("response"),
						},
					},
				},
				&ast.ReturnStmt{},
			},
		},
	})
}

func (h *HandlersFile) AddWriteResponseMethod(baseName string, codes []string) {
	switchBody := &ast.BlockStmt{
		List: []ast.Stmt{},
	}
	for _, code := range codes {
		switchBody.List = append(switchBody.List, &ast.CaseClause{
			List: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.INT,
					Value: code,
				},
			},
			Body: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("response"),
							Sel: ast.NewIdent("Response" + code),
						},
						Op: token.EQL,
						Y:  ast.NewIdent("nil"),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   ast.NewIdent("http"),
										Sel: ast.NewIdent("Error"),
									},
									Args: []ast.Expr{
										ast.NewIdent("w"),
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: `"InternalServerError"`,
										},
										&ast.SelectorExpr{
											X:   ast.NewIdent("http"),
											Sel: ast.NewIdent("StatusInternalServerError"),
										},
									},
								},
							},
							&ast.ReturnStmt{},
						},
					},
				},

				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("h"),
							Sel: ast.NewIdent("write" + baseName + code + "Response"),
						},
						Args: []ast.Expr{
							ast.NewIdent("w"),
							&ast.SelectorExpr{
								X:   ast.NewIdent("response"),
								Sel: ast.NewIdent("Response" + code),
							},
						},
					},
				},
			},
		})
	}
	h.restDecls = append(h.restDecls, &ast.FuncDecl{
		Name: ast.NewIdent("write" + baseName + "Response"),
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("h")},
				Type: &ast.StarExpr{
					X: ast.NewIdent("Handler"),
				},
			}},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("w")},
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("http"),
						Sel: ast.NewIdent("ResponseWriter"),
					},
				}, {
					Names: []*ast.Ident{ast.NewIdent("response")},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("models"),
							Sel: ast.NewIdent(baseName + "Response"),
						},
					},
				}},
			},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.SwitchStmt{
					Tag: &ast.SelectorExpr{
						X:   ast.NewIdent("response"),
						Sel: ast.NewIdent("StatusCode"),
					},
					Body: switchBody,
				},
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("w"),
							Sel: ast.NewIdent("WriteHeader"),
						},
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   ast.NewIdent("response"),
								Sel: ast.NewIdent("StatusCode"),
							},
						},
					},
				},
			},
		},
	})
}

func (h *HandlersFile) AddWriteResponseCode(baseName string, code string, response *openapi3.ResponseRef) error {
	var body []ast.Stmt

	if len(response.Value.Headers) > 0 {
		h.AddImport("encoding/json")
		body = append(body, &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("headersJSON"),
				ast.NewIdent("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("json"),
						Sel: ast.NewIdent("Marshal"),
					},
					Args: []ast.Expr{
						ast.NewIdent("h"),
					},
				},
			},
		})
		body = append(body, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("http"),
								Sel: ast.NewIdent("Error"),
							},
							Args: []ast.Expr{
								ast.NewIdent("w"),
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: `"InternalServerError"`,
								},
								&ast.SelectorExpr{
									X:   ast.NewIdent("http"),
									Sel: ast.NewIdent("StatusInternalServerError"),
								},
							},
						},
					},
					&ast.ReturnStmt{},
				},
			},
		})
		body = append(body, &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent("headers")},
						Type: &ast.MapType{
							Key:   ast.NewIdent("string"),
							Value: ast.NewIdent("string"),
						},
					},
				},
			},
		})
		body = append(body, &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("err")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("json"),
						Sel: ast.NewIdent("Unmarshal"),
					},
					Args: []ast.Expr{
						ast.NewIdent("headersJSON"),
						&ast.UnaryExpr{
							Op: token.AND,
							X:  ast.NewIdent("headers"),
						},
					},
				},
			},
		})
		body = append(body, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("http"),
								Sel: ast.NewIdent("Error"),
							},
							Args: []ast.Expr{
								ast.NewIdent("w"),
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: `"InternalServerError"`,
								},
								&ast.SelectorExpr{
									X:   ast.NewIdent("http"),
									Sel: ast.NewIdent("StatusInternalServerError"),
								},
							},
						},
					},
					&ast.ReturnStmt{},
				},
			},
		})
		body = append(body, &ast.RangeStmt{
			Key:   ast.NewIdent("key"),
			Value: ast.NewIdent("value"),
			Tok:   token.DEFINE,
			X:     ast.NewIdent("headers"),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   ast.NewIdent("w"),
										Sel: ast.NewIdent("Header"),
									},
									Args: []ast.Expr{},
								},
								Sel: ast.NewIdent("Set"),
							},
							Args: []ast.Expr{
								ast.NewIdent("key"),
								ast.NewIdent("value"),
							},
						},
					},
				},
			},
		})
	}

	/*
		    err := json.NewEncoder(w).Encode(r.Body)
			if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}*/
	if len(response.Value.Content) > 1 {
		return errors.New("multiple responses are not supported")
	}
	for key, value := range response.Value.Content {
		if key != applicationJSONCT {
			return errors.New("only application/json content type is supported")
		}
		if value.Schema != nil {
			h.AddImport("encoding/json")
			body = append(body, &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("err")},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("json"),
									Sel: ast.NewIdent("NewEncoder"),
								},
								Args: []ast.Expr{ast.NewIdent("w")},
							},
							Sel: ast.NewIdent("Encode"),
						},
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   ast.NewIdent("r"),
								Sel: ast.NewIdent("Body"),
							},
						},
					},
				},
			})
			body = append(body, &ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent("err"),
					Op: token.NEQ,
					Y:  ast.NewIdent("nil"),
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("http"),
									Sel: ast.NewIdent("Error"),
								},
								Args: []ast.Expr{
									ast.NewIdent("w"),
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: `"InternalServerError"`,
									},
									&ast.SelectorExpr{
										X:   ast.NewIdent("http"),
										Sel: ast.NewIdent("StatusInternalServerError"),
									},
								},
							},
						},
						&ast.ReturnStmt{},
					},
				},
			})
		}
	}

	if len(body) > 0 {
		body = append([]ast.Stmt{&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent("err")},
						Type:  ast.NewIdent("error"),
					},
				},
			},
		}}, body...)
	}

	h.restDecls = append(h.restDecls, &ast.FuncDecl{
		Name: ast.NewIdent("write" + baseName + code + "Response"),
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Names: []*ast.Ident{ast.NewIdent("h")},
				Type: &ast.StarExpr{
					X: ast.NewIdent("Handler"),
				},
			}},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{{
					Names: []*ast.Ident{ast.NewIdent("w")},
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("http"),
						Sel: ast.NewIdent("ResponseWriter"),
					},
				}, {
					Names: []*ast.Ident{ast.NewIdent("r")},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("models"),
							Sel: ast.NewIdent(baseName + "Response" + code),
						},
					},
				}},
			},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: body,
		},
	})

	/*func (h *Handler) WriteGetHelloJson200Response(w http.ResponseWriter, r *models.GetHelloJsonResponse200) {
		if r.Body != nil {
			if err := json.NewEncoder(w).Encode(r.Body); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}
		}
		if r.Headers != nil {
			headers, err := r.Headers.Map()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for key, value := range headers {
				w.Header().Set(key, value)
			}
		}
	}*/
	return nil
}
