package cdshooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://cds.example.org")
	require.NotNil(t, client)
	assert.Equal(t, "https://cds.example.org", client.baseURL)
	assert.NotNil(t, client.httpClient)
}

func TestNewClient_WithOptions(t *testing.T) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	client := NewClient("https://cds.example.org",
		WithHTTPClient(httpClient),
		WithBearerToken("test-token"),
		WithTimeout(5*time.Second),
	)

	assert.Equal(t, "test-token", client.bearerToken)
	assert.Equal(t, 5*time.Second, client.httpClient.Timeout)
}

func TestWithHTTPClient(t *testing.T) {
	opt := WithHTTPClient(&http.Client{})
	client := &Client{}
	opt(client)
	assert.NotNil(t, client.httpClient)
}

func TestWithBearerToken(t *testing.T) {
	opt := WithBearerToken("my-token")
	client := &Client{}
	opt(client)
	assert.Equal(t, "my-token", client.bearerToken)
}

func TestWithTimeout(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{},
	}
	opt := WithTimeout(10 * time.Second)
	opt(client)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestWithFHIRServer(t *testing.T) {
	auth := &FHIRAuth{
		AccessToken: "token",
		TokenType:   "Bearer",
	}
	opt := WithFHIRServer("https://fhir.example.org", auth)
	client := &Client{}
	opt(client)

	assert.Equal(t, "https://fhir.example.org", client.fhirServer)
	assert.NotNil(t, client.fhirAuth)
}

func TestClient_Discover(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/cds-services", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]Service{
			"services": {
				{ID: "svc1", Hook: HookPatientView},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	services, err := client.Discover(context.Background())

	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "svc1", services[0].ID)
}

func TestClient_Discover_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Discover(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestClient_Call(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/cds-services/test-svc", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req CDSRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "patient-view", string(req.Hook))
		assert.NotEmpty(t, req.HookInstance)
		assert.NotNil(t, req.Context)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CDSResponse{
			Cards: []Card{
				{Summary: "Test card", Indicator: IndicatorInfo, Source: Source{Label: "Test"}},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.Call(context.Background(), "test-svc",
		PatientViewContext{UserID: "Practitioner/123", PatientID: "Patient/456"},
		nil,
	)

	require.NoError(t, err)
	require.Len(t, resp.Cards, 1)
	assert.Equal(t, "Test card", resp.Cards[0].Summary)
}

func TestClient_Call_WithPrefetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CDSRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.NotNil(t, req.Prefetch)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CDSResponse{Cards: []Card{}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Call(context.Background(), "test-svc",
		PatientViewContext{UserID: "Practitioner/123", PatientID: "Patient/456"},
		map[string]any{
			"patient": map[string]any{"resourceType": "Patient", "id": "123"},
		},
	)

	require.NoError(t, err)
}

func TestClient_Call_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Call(context.Background(), "test-svc",
		PatientViewContext{}, nil,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestClient_Call_WithBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer my-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CDSResponse{Cards: []Card{}})
	}))
	defer server.Close()

	client := NewClient(server.URL, WithBearerToken("my-token"))
	_, err := client.Call(context.Background(), "test-svc",
		PatientViewContext{}, nil,
	)

	require.NoError(t, err)
}

func TestClient_Feedback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/cds-services/test-svc/feedback", r.URL.Path)

		var feedback FeedbackRequest
		json.NewDecoder(r.Body).Decode(&feedback)

		assert.Equal(t, "card-uuid", feedback.Card)
		assert.Equal(t, OutcomeAccepted, feedback.Outcome)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FeedbackResponse{Status: "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Feedback(context.Background(), "test-svc", FeedbackRequest{
		Card:    "card-uuid",
		Outcome: OutcomeAccepted,
	})

	require.NoError(t, err)
}

func TestClient_Feedback_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Feedback(context.Background(), "test-svc", FeedbackRequest{
		Card:    "card-uuid",
		Outcome: OutcomeAccepted,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestGetHookType(t *testing.T) {
	tests := []struct {
		ctx      any
		expected string
	}{
		{PatientViewContext{}, "patient-view"},
		{OrderSelectContext{}, "order-select"},
		{OrderSignContext{}, "order-sign"},
		{AppointmentBookContext{}, "appointment-book"},
		{EncounterStartContext{}, "encounter-start"},
		{EncounterDischargeContext{}, "encounter-discharge"},
		{struct{}{}, ""},
	}

	for _, tt := range tests {
		result := getHookType(tt.ctx)
		assert.Equal(t, tt.expected, result)
	}
}
