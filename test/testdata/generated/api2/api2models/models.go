package api2models

import "github.com/jolfzverb/codegen/internal/generated/generated/def/defmodels"

type CreateRequest struct {
	Body defmodels.NewResourseResponse
}
type CreateResponse200 struct {
}
type CreateResponse struct {
	StatusCode  int
	Response200 *CreateResponse200
}
