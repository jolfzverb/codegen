package generator

import (
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

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

func (g *Generator) ProcessApplicationJSONOperation(pathName string, method string, contentType string,
	_ *openapi3.Operation,
) error {
	const op = "generator.ProcessApplicationJsonOperation"
	suffix, err := NameSuffixFromContentType(contentType)
	if err != nil {
		return errors.Wrap(err, op)
	}
	handlerBaseName := FormatGoLikeIdentifier(method) + FormatGoLikeIdentifier(pathName) + suffix

	g.AddInterface(handlerBaseName)
	g.AddDependencyToHandler(handlerBaseName)
	g.AddRoute(handlerBaseName, method, pathName)
	// if path params add ParsePathParams method
	// if query params add ParseQueryParams method
	// if header params add ParseHeaderParams method
	// if cookie params add ParseCookieParams method
	// if request body add ParseRequestBody method
	// add parse params method
	// add handlejson method
	// add/modify handle method
	// add path params model to models
	// add query params model to models
	// add header params model to models
	// add cookie params model to models
	// add request body model to models
	// add response model to models
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
			case "application/json":
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
	}

	return nil
}
