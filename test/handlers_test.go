package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jolfzverb/codegen/internal/usage/generated/api"
	"github.com/jolfzverb/codegen/internal/usage/generated/api/models"
	"github.com/stretchr/testify/assert"
)

// Mock implementations of the handler interfaces
type mockPostSessionNewJSONHandler struct{}

func (m *mockPostSessionNewJSONHandler) HandlePostPathToParamResourseJSON(ctx context.Context, r *models.PostPathToParamResourseJSONRequest) (*models.PostPathToParamResourseJSONResponse, error) {
	if r.Body.CodeForResponse != nil {
		switch *r.Body.CodeForResponse {
		case 400:
			return &models.PostPathToParamResourseJSONResponse{
				StatusCode:  400,
				Response400: &models.PostPathToParamResourseJSONResponse400{},
			}, nil
		case 404:
			return &models.PostPathToParamResourseJSONResponse{
				StatusCode:  404,
				Response404: &models.PostPathToParamResourseJSONResponse404{},
			}, nil
		}
	}
	return &models.PostPathToParamResourseJSONResponse{
		StatusCode: 200,
		Response200: &models.PostPathToParamResourseJSONResponse200{
			Body: models.PostPathToParamResourseJSONResponse200Body{
				Count:       r.Query.Count,
				Description: r.Body.Description,
				Name:        r.Body.Name,
				Param:       r.Path.Param,
			},
			Headers: &models.PostPathToParamResourseJSONResponse200Headers{
				IdempotencyKey: r.Headers.IdempotencyKey,
			},
		},
	}, nil
}

func TestHandler(t *testing.T) {
	router := chi.NewRouter()
	handler := api.NewHandler(
		&mockPostSessionNewJSONHandler{},
	)
	handler.AddRoutes(router)

	// Create a test server using the chi router
	server := httptest.NewServer(router)
	defer server.Close()

	// Test POST /session/new
	t.Run("200 Success", func(t *testing.T) {
		requestBody := `{"name": "value", "description": "descr"}`
		request, err := http.NewRequest(http.MethodPost, server.URL+"/path/to/param/resourse?count=3", bytes.NewBufferString(requestBody))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "unique-idempotency-key")
		resp, err := http.DefaultClient.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var responseBody map[string]any
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "unique-idempotency-key", resp.Header.Get("Idempotency-Key"))
		assert.Equal(t, "3", responseBody["count"])
		assert.Equal(t, "descr", responseBody["description"])
		assert.Equal(t, "value", responseBody["name"])
	})
	t.Run("404", func(t *testing.T) {
		requestBody := `{"name": "value", "description": "descr", "code_for_response": 404}`
		request, err := http.NewRequest(http.MethodPost, server.URL+"/path/to/param/resourse?count=3", bytes.NewBufferString(requestBody))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "unique-idempotency-key")
		resp, err := http.DefaultClient.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
	t.Run("400 No name", func(t *testing.T) {
		requestBody := `{}`
		request, err := http.NewRequest(http.MethodPost, server.URL+"/path/to/param/resourse?count=3", bytes.NewBufferString(requestBody))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "unique-idempotency-key")
		resp, err := http.DefaultClient.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
