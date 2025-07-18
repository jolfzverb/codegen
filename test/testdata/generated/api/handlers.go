// Code generated by github.com/jolfzverb/codegen; DO NOT EDIT.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-faster/errors"
	"github.com/go-playground/validator/v10"
	"github.com/jolfzverb/codegen/test/testdata/generated/api/apimodels"
)

type CreateHandler interface {
	HandleCreate(ctx context.Context, r apimodels.CreateRequest) (*apimodels.CreateResponse, error)
}
type Handler struct {
	validator *validator.Validate
	create    CreateHandler
}

func NewHandler(create CreateHandler) *Handler {
	return &Handler{validator: validator.New(validator.WithRequiredStructEnabled()), create: create}
}
func (h *Handler) AddRoutes(router chi.Router) {
	router.Post("/path/to/{param}/resourse", h.handleCreate)
}
func (h *Handler) parseCreatePathParams(r *http.Request) (*apimodels.CreatePathParams, error) {
	var pathParams apimodels.CreatePathParams
	param := chi.URLParam(r, "param")
	if param == "" {
		return nil, errors.New("param path param is required")
	}
	pathParams.Param = param
	err := h.validator.Struct(pathParams)
	if err != nil {
		return nil, err
	}
	return &pathParams, nil
}
func (h *Handler) parseCreateQueryParams(r *http.Request) (*apimodels.CreateQueryParams, error) {
	var queryParams apimodels.CreateQueryParams
	count := r.URL.Query().Get("count")
	if count == "" {
		return nil, errors.New("count query param is required")
	}
	queryParams.Count = count
	err := h.validator.Struct(queryParams)
	if err != nil {
		return nil, err
	}
	return &queryParams, nil
}
func (h *Handler) parseCreateHeaders(r *http.Request) (*apimodels.CreateHeaders, error) {
	var headers apimodels.CreateHeaders
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return nil, errors.New("Idempotency-Key header is required")
	}
	headers.IdempotencyKey = idempotencyKey
	optionalHeader := r.Header.Get("Optional-Header")
	if optionalHeader != "" {
		parsedOptionalHeader, err := time.Parse(time.RFC3339, optionalHeader)
		if err != nil {
			return nil, errors.Wrap(err, "OptionalHeader is not a valid date-time format")
		}
		headers.OptionalHeader = &parsedOptionalHeader
	}
	err := h.validator.Struct(headers)
	if err != nil {
		return nil, err
	}
	return &headers, nil
}
func (h *Handler) parseCreateCookies(r *http.Request) (*apimodels.CreateCookies, error) {
	var cookies apimodels.CreateCookies
	cookieParam, err := r.Cookie("cookie-param")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return nil, err
	}
	if err == nil {
		cookieParamValue := cookieParam.Value
		cookies.CookieParam = &cookieParamValue
	}
	requiredCookieParam, err := r.Cookie("required-cookie-param")
	if err != nil {
		return nil, err
	}
	requiredCookieParamValue := requiredCookieParam.Value
	cookies.RequiredCookieParam = requiredCookieParamValue
	err = h.validator.Struct(cookies)
	if err != nil {
		return nil, err
	}
	return &cookies, nil
}
func ValidateCreateRequestBodyObjectArrayItemJSON(_ json.RawMessage) error {
	return nil
}
func ValidateCreateRequestBodyObjectArrayJSON(jsonData json.RawMessage) error {
	var arr []json.RawMessage
	err := json.Unmarshal(jsonData, &arr)
	if err != nil {
		return err
	}
	for index, obj := range arr {
		if !containsNull(obj) {
			err = ValidateCreateRequestBodyObjectArrayItemJSON(obj)
			if err != nil {
				return errors.Wrapf(err, "error validating object at index %d", index)
			}
		}
	}
	return nil
}
func ValidateCreateRequestBodyObjectFieldField2JSON(_ json.RawMessage) error {
	return nil
}
func containsNull(data json.RawMessage) bool {
	var temp any
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return false
	}
	return temp == nil
}
func ValidateCreateRequestBodyObjectFieldJSON(jsonData json.RawMessage) error {
	var obj map[string]json.RawMessage
	err := json.Unmarshal(jsonData, &obj)
	if err != nil {
		return err
	}
	var val json.RawMessage
	var exists bool
	val, exists = obj["field2"]
	if exists && !containsNull(val) {
		err = ValidateCreateRequestBodyObjectFieldField2JSON(val)
		if err != nil {
			return errors.Wrap(err, "field field2 is not valid")
		}
	}
	return nil
}
func ValidateCreateRequestBodyJSON(jsonData json.RawMessage) error {
	requiredFields := map[string]bool{"name": true}
	nullableFields := map[string]bool{}
	var obj map[string]json.RawMessage
	err := json.Unmarshal(jsonData, &obj)
	if err != nil {
		return err
	}
	var val json.RawMessage
	var exists bool
	for field := range requiredFields {
		val, exists = obj[field]
		if !exists {
			return errors.New("field " + field + " is required")
		}
		if !nullableFields[field] && containsNull(val) {
			return errors.New("field " + field + " cannot be null")
		}
	}
	val, exists = obj["object-array"]
	if exists && !containsNull(val) {
		err = ValidateCreateRequestBodyObjectArrayJSON(val)
		if err != nil {
			return errors.Wrap(err, "field object-array is not valid")
		}
	}
	val, exists = obj["object-field"]
	if exists && !containsNull(val) {
		err = ValidateCreateRequestBodyObjectFieldJSON(val)
		if err != nil {
			return errors.Wrap(err, "field object-field is not valid")
		}
	}
	return nil
}
func (h *Handler) parseCreateRequestBody(r *http.Request) (*apimodels.CreateRequestBody, error) {
	var bodyJSON json.RawMessage
	err := json.NewDecoder(r.Body).Decode(&bodyJSON)
	if err != nil {
		return nil, err
	}
	err = ValidateCreateRequestBodyJSON(bodyJSON)
	if err != nil {
		return nil, err
	}
	var body apimodels.CreateRequestBody
	err = json.Unmarshal(bodyJSON, &body)
	if err != nil {
		return nil, err
	}
	err = h.validator.Struct(body)
	if err != nil {
		return nil, err
	}
	return &body, nil
}
func (h *Handler) parseCreateRequest(r *http.Request) (*apimodels.CreateRequest, error) {
	pathParams, err := h.parseCreatePathParams(r)
	if err != nil {
		return nil, err
	}
	queryParams, err := h.parseCreateQueryParams(r)
	if err != nil {
		return nil, err
	}
	headers, err := h.parseCreateHeaders(r)
	if err != nil {
		return nil, err
	}
	cookieParams, err := h.parseCreateCookies(r)
	if err != nil {
		return nil, err
	}
	body, err := h.parseCreateRequestBody(r)
	if err != nil {
		return nil, err
	}
	return &apimodels.CreateRequest{Path: *pathParams, Query: *queryParams, Headers: *headers, Cookies: *cookieParams, Body: *body}, nil
}
func Create200Response(body apimodels.NewResourseResponse, headers apimodels.CreateResponse200Headers) *apimodels.CreateResponse {
	return &apimodels.CreateResponse{StatusCode: 200, Response200: &apimodels.CreateResponse200{Body: body, Headers: headers}}
}
func (h *Handler) writeCreate200Response(w http.ResponseWriter, r *apimodels.CreateResponse200) {
	var err error
	headersJSON, err := json.Marshal(r.Headers)
	if err != nil {
		http.Error(w, "{\"error\":\"InternalServerError\"}", http.StatusInternalServerError)
		return
	}
	var headers map[string]string
	err = json.Unmarshal(headersJSON, &headers)
	if err != nil {
		http.Error(w, "{\"error\":\"InternalServerError\"}", http.StatusInternalServerError)
		return
	}
	for key, value := range headers {
		w.Header().Set(key, value)
	}
	err = json.NewEncoder(w).Encode(r.Body)
	if err != nil {
		http.Error(w, "{\"error\":\"InternalServerError\"}", http.StatusInternalServerError)
		return
	}
}
func Create400Response() *apimodels.CreateResponse {
	return &apimodels.CreateResponse{StatusCode: 400, Response400: &apimodels.CreateResponse400{}}
}
func (h *Handler) writeCreate400Response(w http.ResponseWriter, r *apimodels.CreateResponse400) {
}
func Create404Response() *apimodels.CreateResponse {
	return &apimodels.CreateResponse{StatusCode: 404, Response404: &apimodels.CreateResponse404{}}
}
func (h *Handler) writeCreate404Response(w http.ResponseWriter, r *apimodels.CreateResponse404) {
}
func (h *Handler) writeCreateResponse(w http.ResponseWriter, response *apimodels.CreateResponse) {
	switch response.StatusCode {
	case 200:
		if response.Response200 == nil {
			http.Error(w, "{\"error\":\"InternalServerError\"}", http.StatusInternalServerError)
			return
		}
		h.writeCreate200Response(w, response.Response200)
	case 400:
		if response.Response400 == nil {
			http.Error(w, "{\"error\":\"InternalServerError\"}", http.StatusInternalServerError)
			return
		}
		h.writeCreate400Response(w, response.Response400)
	case 404:
		if response.Response404 == nil {
			http.Error(w, "{\"error\":\"InternalServerError\"}", http.StatusInternalServerError)
			return
		}
		h.writeCreate404Response(w, response.Response404)
	}
	w.WriteHeader(response.StatusCode)
}
func (h *Handler) handleCreateRequest(w http.ResponseWriter, r *http.Request) {
	request, err := h.parseCreateRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("{\"error\":%s}", strconv.Quote(err.Error())), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	response, err := h.create.HandleCreate(ctx, *request)
	if err != nil || response == nil {
		http.Error(w, "{\"error\":\"InternalServerError\"}", http.StatusInternalServerError)
		return
	}
	h.writeCreateResponse(w, response)
	return
}
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		h.handleCreateRequest(w, r)
		return
	case "":
		h.handleCreateRequest(w, r)
		return
	default:
		http.Error(w, "{\"error\":\"Unsupported Content-Type\"}", http.StatusUnsupportedMediaType)
		return
	}
}
func ValidateNewResourseResponseJSON(jsonData json.RawMessage) error {
	requiredFields := map[string]bool{"count": true, "name": true, "param": true}
	nullableFields := map[string]bool{}
	var obj map[string]json.RawMessage
	err := json.Unmarshal(jsonData, &obj)
	if err != nil {
		return err
	}
	var val json.RawMessage
	var exists bool
	for field := range requiredFields {
		val, exists = obj[field]
		if !exists {
			return errors.New("field " + field + " is required")
		}
		if !nullableFields[field] && containsNull(val) {
			return errors.New("field " + field + " cannot be null")
		}
	}
	return nil
}
