package cdshooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_Discovery(t *testing.T) {
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "test-service",
			Hook:        HookPatientView,
			Title:       "Test Service",
			Description: "A test service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string][]Service
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result["services"], 1)
	assert.Equal(t, "test-service", result["services"][0].ID)
}

func TestServer_Discovery_Empty(t *testing.T) {
	server := NewServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string][]Service
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result["services"], 0)
}

func TestServer_UnknownService(t *testing.T) {
	server := NewServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/unknown", nil)
	r.Header.Set("Content-Type", "application/json")
	r.Body = nil
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServer_CORS(t *testing.T) {
	server := NewServer(WithCORSOrigins("*"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/cds-services", nil)
	r.Header.Set("Origin", "http://example.com")
	r.Header.Set("Access-Control-Request-Method", "POST")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestServer_HookInstanceValidation(t *testing.T) {
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "test-service",
			Hook:        HookPatientView,
			Title:       "Test Service",
			Description: "A test service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}),
	})

	body := `{"hook":"patient-view","context":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "hookInstance")
}

func TestServer_CardsNeverNull(t *testing.T) {
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "test-service",
			Hook:        HookPatientView,
			Title:       "Test Service",
			Description: "A test service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}),
	})

	body := `{"hook":"patient-view","hookInstance":"550e8400-e29b-41d4-a716-446655440000","context":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Cards)
}
