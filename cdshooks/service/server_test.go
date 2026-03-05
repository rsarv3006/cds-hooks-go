package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	cdshooks "github.com/your-org/cds-hooks-go/cdshooks"
)

func TestServer_Discovery(t *testing.T) {
	title := "Test Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
		}),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string][]cdshooks.Service
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

	var result map[string][]cdshooks.Service
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
	title := "Test Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
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
	title := "Test Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
		}),
	})

	body := `{"hook":"patient-view","hookInstance":"550e8400-e29b-41d4-a716-446655440000","context":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp cdshooks.CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Cards)
}

func TestServer_FeedbackEndpoint(t *testing.T) {
	title := "Test Service"
	feedbackReceived := false
	var receivedServiceID string
	var receivedFeedback cdshooks.FeedbackRequest

	server := NewServer(WithFeedbackHandler(&testFeedbackHandler{
		fn: func(ctx context.Context, serviceID string, feedback cdshooks.FeedbackRequest) error {
			feedbackReceived = true
			receivedServiceID = serviceID
			receivedFeedback = feedback
			return nil
		},
	}))
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
		}),
	})

	feedbackBody := `{"card":"card-123","outcome":"accepted","acceptedSuggestions":[{"id":"suggestion-456"}],"outcomeTimestamp":"2024-01-01T00:00:00Z"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service/feedback",
		strings.NewReader(feedbackBody))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, feedbackReceived)
	assert.Equal(t, "test-service", receivedServiceID)
	assert.Equal(t, "card-123", receivedFeedback.Card)
	assert.Equal(t, cdshooks.OutcomeAccepted, receivedFeedback.Outcome)
}

func TestServer_FeedbackEndpoint_Override(t *testing.T) {
	title := "Test Service"
	var receivedFeedback cdshooks.FeedbackRequest

	server := NewServer(WithFeedbackHandler(&testFeedbackHandler{
		fn: func(ctx context.Context, serviceID string, feedback cdshooks.FeedbackRequest) error {
			receivedFeedback = feedback
			return nil
		},
	}))
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
		}),
	})

	reason := cdshooks.OverrideReason{
		Reason: &cdshooks.Coding{Code: "reason-code"},
	}
	reasonJSON, _ := json.Marshal(reason)
	feedbackBody := `{"card":"card-123","outcome":"overridden","overrideReason":` + string(reasonJSON) + `,"outcomeTimestamp":"2024-01-01T00:00:00Z"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service/feedback",
		strings.NewReader(feedbackBody))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, cdshooks.OutcomeOverridden, receivedFeedback.Outcome)
	assert.NotNil(t, receivedFeedback.OverrideReason)
}

func TestServer_FeedbackEndpoint_NotEnabled(t *testing.T) {
	server := NewServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service/feedback",
		strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServer_FeedbackEndpoint_InvalidBody(t *testing.T) {
	title := "Test Service"
	server := NewServer(WithFeedbackHandler(&testFeedbackHandler{
		fn: func(ctx context.Context, serviceID string, feedback cdshooks.FeedbackRequest) error {
			return nil
		},
	}))
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
		}),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service/feedback",
		strings.NewReader(`{invalid`))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

type testFeedbackHandler struct {
	fn func(ctx context.Context, serviceID string, feedback cdshooks.FeedbackRequest) error
}

func (h *testFeedbackHandler) Feedback(ctx context.Context, serviceID string, feedback cdshooks.FeedbackRequest) error {
	return h.fn(ctx, serviceID, feedback)
}

func TestServer_ErrorHandling_InvalidJSON(t *testing.T) {
	title := "Test Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
		}),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service",
		strings.NewReader(`{invalid json`))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestServer_ErrorHandling_MissingHookInstance(t *testing.T) {
	title := "Test Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
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

func TestServer_ErrorHandling_InvalidHookInstance(t *testing.T) {
	title := "Test Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
		}),
	})

	body := `{"hook":"patient-view","hookInstance":"not-a-uuid","context":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "UUID")
}

func TestServer_ErrorHandling_404(t *testing.T) {
	server := NewServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/nonexistent",
		strings.NewReader(`{"hookInstance":"550e8400-e29b-41d4-a716-446655440000"}`))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not found")
}
