// Code generated by github.com/jolfzverb/codegen; DO NOT EDIT.

package api3models

import "github.com/jolfzverb/codegen/test/testdata/generated/def/defmodels"

type CreateRequest struct {
}
type CreateResponse200 struct {
	Body defmodels.NewResourseResponse
}
type CreateResponse struct {
	StatusCode  int
	Response200 *CreateResponse200
}
