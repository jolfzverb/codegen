package generator

import (
	"context"
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

func (g *Generator) AddInterface(ctx context.Context, baseName string) {
	const op = "generator.AddInterface"
	interfaceName := baseName + "Handler"
	methodName := "Handle" + baseName
	requestName := baseName + "Request"
	responseName := baseName + "Response"
	g.HandlersFile.AddInterface(interfaceName, methodName, requestName, responseName)
}

func (g *Generator) AddDependencyToHandler(ctx context.Context, baseName string) {
	const op = "generator.AddDependencyToHandler"

	g.HandlersFile.AddDependencyToHandler(baseName)
}

func (g *Generator) AddRoute(ctx context.Context, baseName string, method string, pathName string) {
	const op = "generator.AddRoute"
	g.HandlersFile.AddRouteToRouter(baseName, method, pathName)
}

func (g *Generator) ProcessApplicationJsonOperation(ctx context.Context, pathName string, method string, contentType string, operation *openapi3.Operation) error {
	const op = "generator.ProcessApplicationJsonOperation"
	suffix, err := NameSuffixFromContentType(contentType)
	if err != nil {
		return errors.Wrap(err, op)
	}
	handlerBaseName := FormatGoLikeIdentifier(method) + FormatGoLikeIdentifier(pathName) + suffix

	g.AddInterface(ctx, handlerBaseName)
	g.AddDependencyToHandler(ctx, handlerBaseName)
	g.AddRoute(ctx, handlerBaseName, method, pathName)
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

func (g *Generator) ProcessOperation(ctx context.Context, pathName string, method string, operation *openapi3.Operation) error {
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
				err := g.ProcessApplicationJsonOperation(ctx, pathName, method, contentType, operation)
				if err != nil {
					return errors.Wrap(err, op)
				}
			default:
				return fmt.Errorf("unsupported content type %s for operation %s %s", contentType, method, pathName)
			}
		}
	} else {
		err := g.ProcessApplicationJsonOperation(ctx, pathName, method, "", operation)
		if err != nil {
			return errors.Wrap(err, op)
		}
	}
	return nil
}

func (g *Generator) ProcessPaths(ctx context.Context, paths *openapi3.Paths) error {
	const op = "generator.ProcessPaths"
	for _, pathName := range paths.InMatchingOrder() {
		pathItem := paths.Value(pathName)
		if pathItem.Get != nil {
			if pathItem.Get.RequestBody != nil {
				return fmt.Errorf("GET method should not have request body")
			}
			g.ProcessOperation(ctx, pathName, "Get", pathItem.Get)
		}
	}
	return nil
}
