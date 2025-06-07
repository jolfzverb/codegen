package generator

import (
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

const applicationJSONCT = "application/json"

func (g *Generator) AddInterface(baseName string) {
	interfaceName := baseName + "Handler"
	methodName := "Handle" + baseName
	requestName := baseName + "Request"
	responseName := baseName + "Response"
	g.HandlersFile.AddInterface(interfaceName, methodName, requestName, responseName)
}

func (g *Generator) AddDependencyToHandler(baseName string) {
	g.HandlersFile.AddDependencyToHandler(baseName)
}

func (g *Generator) AddRoute(baseName string, method string, pathName string) {
	g.HandlersFile.AddRouteToRouter(baseName, method, pathName)
}

func (g *Generator) AddContentTypeToHandler(baseName string, rawContentType string) {
	if g.HandlersFile.GetHandler(baseName) == nil {
		g.HandlersFile.CreateHandler(baseName)
	}
	g.HandlersFile.AddContentTypeHandler(baseName, rawContentType)
}

func (g *Generator) AddHandleOperationMethod(baseName string) {
	g.HandlersFile.AddHandleOperationMethod(baseName)
}

func (g *Generator) AddResponseCodeModels(baseName string, code string, response *openapi3.ResponseRef) error {
	const op = "generator.AddResponseCodeModels"
	if len(response.Value.Content) > 1 {
		return errors.New("multiple response content types are not supported")
	}
	model := SchemaStruct{
		Name:   baseName + "Response" + code,
		Fields: []SchemaField{},
	}
	for _, content := range response.Value.Content {
		if content.Schema != nil {
			if content.Schema.Ref == "" {
				err := g.SchemasFile.ProcessObjectSchema(baseName+"Response"+code+"Body", content.Schema)
				if err != nil {
					return errors.Wrap(err, op)
				}
			}
			typeName := baseName + "Response" + code + "Body"
			if content.Schema.Ref != "" {
				typeName = ParseRefTypeName(content.Schema.Ref)
			}
			model.Fields = append(model.Fields, SchemaField{
				Name:        "Body",
				Type:        typeName,
				TagJSON:     []string{},
				TagValidate: []string{},
				Required:    true,
			})
		}
	}
	if len(response.Value.Headers) > 0 {
		err := g.SchemasFile.AddHeadersModel(baseName+"Response"+code, response.Value.Headers)
		if err != nil {
			return errors.Wrap(err, op)
		}
		model.Fields = append(model.Fields, SchemaField{
			Name:     "Headers",
			Type:     baseName + "Response" + code + "Headers",
			Required: true,
		})
	}
	g.SchemasFile.AddSchema(model)
	err := g.HandlersFile.AddCreateResponseModel(baseName, code, response)
	if err != nil {
		return errors.Wrapf(err, op)
	}

	return nil
}

func (g *Generator) AddResponseModel(baseName string, responseCodes []string) {
	model := SchemaStruct{
		Name: baseName + "Response",
		Fields: []SchemaField{
			{
				Name:     "StatusCode",
				Type:     "int",
				Required: true,
			},
		},
	}
	for _, code := range responseCodes {
		field := SchemaField{
			Name: "Response" + code,
			Type: baseName + "Response" + code,
		}
		model.Fields = append(model.Fields, field)
	}
	g.SchemasFile.AddSchema(model)
}

func (g *Generator) AddWriteResponseMethod(baseName string, operation *openapi3.Operation) error {
	const op = "generator.AddWriteResponseMethod"
	var err error
	codes := make([]string, 0, len(operation.Responses.Map()))
	keys := make([]string, 0, len(operation.Responses.Map()))
	for code := range operation.Responses.Map() {
		keys = append(keys, code)
	}
	sort.Strings(keys)
	for _, code := range keys {
		response := operation.Responses.Value(code)
		err = g.AddResponseCodeModels(baseName, code, response)
		if err != nil {
			return errors.Wrapf(err, op)
		}
		err = g.HandlersFile.AddWriteResponseCode(baseName, code, response)
		if err != nil {
			return errors.Wrapf(err, op)
		}
		codes = append(codes, code)
	}
	g.HandlersFile.AddWriteResponseMethod(baseName, codes)
	g.AddResponseModel(baseName, keys)

	return nil
}

func (g *Generator) GetOperationParamsByType(operation *openapi3.Operation, paramIn string) openapi3.Parameters {
	var result openapi3.Parameters
	for _, p := range operation.Parameters {
		if p.Value.In == paramIn {
			result = append(result, p)
		}
	}

	return result
}

func (g *Generator) AddParseParamsMethods(baseName string, contentType string, operation *openapi3.Operation) error {
	const op = "generator.AddParseParamsMethods"
	var err error

	pathParams := g.GetOperationParamsByType(operation, openapi3.ParameterInPath)
	if len(pathParams) > 0 {
		err = g.SchemasFile.AddParamsModel(baseName, "PathParams", pathParams)
		if err != nil {
			return errors.Wrap(err, op)
		}
		err = g.HandlersFile.AddParsePathParamsMethod(baseName, pathParams)
		if err != nil {
			return errors.Wrap(err, op)
		}
	}
	queryParams := g.GetOperationParamsByType(operation, openapi3.ParameterInQuery)
	if len(queryParams) > 0 {
		err = g.SchemasFile.AddParamsModel(baseName, "QueryParams", queryParams)
		if err != nil {
			return errors.Wrap(err, op)
		}
		err = g.HandlersFile.AddParseQueryParamsMethod(baseName, queryParams)
		if err != nil {
			return errors.Wrap(err, op)
		}
	}
	headerParams := g.GetOperationParamsByType(operation, openapi3.ParameterInHeader)
	if len(headerParams) > 0 {
		err = g.SchemasFile.AddParamsModel(baseName, "Headers", headerParams)
		if err != nil {
			return errors.Wrap(err, op)
		}
		err = g.HandlersFile.AddParseHeadersMethod(baseName, headerParams)
		if err != nil {
			return errors.Wrap(err, op)
		}
	}
	cookieParams := g.GetOperationParamsByType(operation, openapi3.ParameterInCookie)
	if len(cookieParams) > 0 {
		err = g.SchemasFile.AddParamsModel(baseName, "Cookies", cookieParams)
		if err != nil {
			return errors.Wrap(err, op)
		}
		err = g.HandlersFile.AddParseCookiesMethod(baseName, cookieParams)
		if err != nil {
			return errors.Wrap(err, op)
		}
	}
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		content, ok := operation.RequestBody.Value.Content[contentType]
		if ok && content.Schema != nil {
			if content.Schema.Ref == "" {
				err = g.SchemasFile.ProcessObjectSchema(baseName+"RequestBody", content.Schema)
				if err != nil {
					return errors.Wrap(err, op)
				}
			}
			err = g.HandlersFile.AddParseRequestBodyMethod(baseName, contentType, operation.RequestBody)
			if err != nil {
				return errors.Wrap(err, op)
			}
		}
	}
	g.HandlersFile.AddParseRequestMethod(baseName, contentType,
		pathParams, queryParams, headerParams, cookieParams, operation.RequestBody,
	)
	g.SchemasFile.GenerateRequestModel(baseName, contentType,
		pathParams, queryParams, headerParams, cookieParams, operation.RequestBody,
	)

	return nil
}

func (g *Generator) ProcessApplicationJSONOperation(pathName string, method string, contentType string,
	operation *openapi3.Operation,
) error {
	const op = "generator.ProcessApplicationJsonOperation"
	if contentType == "" {
		contentType = applicationJSONCT
	}
	handlerBaseName := FormatGoLikeIdentifier(method) + FormatGoLikeIdentifier(pathName)
	if operation.OperationID != "" {
		handlerBaseName = FormatGoLikeIdentifier(operation.OperationID)
	}

	g.AddInterface(handlerBaseName)
	g.AddDependencyToHandler(handlerBaseName)
	g.AddRoute(handlerBaseName, method, pathName)
	err := g.AddParseParamsMethods(handlerBaseName, contentType, operation)
	if err != nil {
		return errors.Wrap(err, op)
	}
	err = g.AddWriteResponseMethod(handlerBaseName, operation)
	if err != nil {
		return errors.Wrap(err, op)
	}
	g.AddHandleOperationMethod(handlerBaseName)
	g.AddContentTypeToHandler(handlerBaseName, contentType)

	return nil
}

func (g *Generator) ProcessOperation(pathName string, method string, operation *openapi3.Operation) error {
	const op = "generator.ProcessOperation"

	if operation.RequestBody != nil {
		contentKeys := make([]string, 0, len(operation.RequestBody.Value.Content))
		for contentType := range operation.RequestBody.Value.Content {
			contentKeys = append(contentKeys, contentType)
		}
		sort.Strings(contentKeys)
		for _, contentType := range contentKeys {
			switch contentType {
			case applicationJSONCT:
				err := g.ProcessApplicationJSONOperation(pathName, method, contentType, operation)
				if err != nil {
					return errors.Wrap(err, op)
				}
			default:
				return errors.New("unsupported content type")
			}
		}
	} else {
		err := g.ProcessApplicationJSONOperation(pathName, method, "", operation)
		if err != nil {
			return errors.Wrap(err, op)
		}
	}

	return nil
}

func (g *Generator) ProcessPaths(paths *openapi3.Paths) error {
	const op = "generator.ProcessPaths"
	for _, pathName := range paths.InMatchingOrder() {
		pathItem := paths.Value(pathName)
		if pathItem.Get != nil {
			if pathItem.Get.RequestBody != nil {
				return errors.New("GET method should not have request body")
			}
			err := g.ProcessOperation(pathName, "Get", pathItem.Get)
			if err != nil {
				return errors.Wrap(err, op)
			}
		}
		if pathItem.Post != nil {
			err := g.ProcessOperation(pathName, "Post", pathItem.Post)
			if err != nil {
				return errors.Wrap(err, op)
			}
		}
		if pathItem.Delete != nil {
			if pathItem.Delete.RequestBody != nil {
				return errors.New("DELETE method should not have request body")
			}
			err := g.ProcessOperation(pathName, "Delete", pathItem.Delete)
			if err != nil {
				return errors.Wrap(err, op)
			}
		}
		if pathItem.Put != nil {
			err := g.ProcessOperation(pathName, "Put", pathItem.Put)
			if err != nil {
				return errors.Wrap(err, op)
			}
		}
		if pathItem.Patch != nil {
			err := g.ProcessOperation(pathName, "Patch", pathItem.Patch)
			if err != nil {
				return errors.Wrap(err, op)
			}
		}
	}

	return nil
}
