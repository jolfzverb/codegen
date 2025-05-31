package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jolfzverb/codegen/internal/usage/generated/api"
	"github.com/jolfzverb/codegen/internal/usage/generated/api/models"
	"github.com/stretchr/testify/assert"
)

type mockHandler struct{}

func (m *mockHandler) HandlePostPathToParamResourseJSON(ctx context.Context, r *models.PostPathToParamResourseJSONRequest) (*models.PostPathToParamResourseJSONResponse, error) {
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
	var date *time.Time
	if r.Body.Date != nil {
		date = new(time.Time)
		*date = r.Body.Date.UTC()
	}
	var date2 *time.Time
	if r.Headers.OptionalHeader != nil {
		date2 = new(time.Time)
		*date2 = r.Headers.OptionalHeader.UTC()
	}
	return &models.PostPathToParamResourseJSONResponse{
		StatusCode: 200,
		Response200: &models.PostPathToParamResourseJSONResponse200{
			Body: models.PostPathToParamResourseJSONResponse200Body{
				Count:       r.Query.Count,
				Description: r.Body.Description,
				Name:        r.Body.Name,
				Param:       r.Path.Param,
				Date:        date,
				Date2:       date2,
				EnumVal:     r.Body.EnumVal,
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
		&mockHandler{},
	)
	handler.AddRoutes(router)

	// Create a test server using the chi router
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("200 Success", func(t *testing.T) {
		requestBody := `{"name": "value", "description": "descr", "date": "2023-10-01T00:00:00+03:00", "code_for_response": 200, "enum-val": "value1"}`
		request, err := http.NewRequest(http.MethodPost, server.URL+"/path/to/param/resourse?count=3", bytes.NewBufferString(requestBody))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "unique-idempotency-key")
		request.Header.Set("Optional-Header", "2023-10-01T00:00:00+03:00")
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
		assert.Equal(t, "2023-09-30T21:00:00Z", responseBody["date"])
		assert.Equal(t, "2023-09-30T21:00:00Z", responseBody["date2"])
		assert.Equal(t, "value1", responseBody["enum-val"])
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
