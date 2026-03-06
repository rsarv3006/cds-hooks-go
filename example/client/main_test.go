package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/cds-hooks-go/cdshooks"
)

func TestClient_Discovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/cds-services", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]cdshooks.Service{
			"services": {
				{
					ID:          "patient-view-age-check",
					Hook:        cdshooks.HookPatientView,
					Title:       ptr("Age Check"),
					Description: "Checks patient age",
					Prefetch:    map[string]string{"patient": "Patient/{{context.patientId}}"},
				},
			},
		})
	}))
	defer server.Close()

	ctx := context.Background()
	client := cdshooks.NewClient(server.URL)
	services, err := client.Discover(ctx)
	require.NoError(t, err)
	require.Len(t, services, 1)
	assert.Equal(t, "patient-view-age-check", services[0].ID)
}

func TestClient_Call_WithPrefetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/cds-services/patient-view-age-check", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req cdshooks.CDSRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "patient-view", string(req.Hook))
		assert.NotEmpty(t, req.HookInstance)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cdshooks.CDSResponse{
			Cards: []cdshooks.Card{
				{
					Summary:   "Patient is 65+",
					Indicator: cdshooks.IndicatorInfo,
					Source:    cdshooks.Source{Label: "Test"},
				},
			},
		})
	}))
	defer server.Close()

	ctx := context.Background()
	client := cdshooks.NewClient(server.URL)
	resp, err := client.Call(ctx, "patient-view-age-check",
		cdshooks.PatientViewContext{
			UserID:    "Practitioner/abc",
			PatientID: "Patient/123",
		},
		map[string]any{
			"patient": map[string]any{
				"resourceType": "Patient",
				"id":           "123",
				"birthDate":    "1955-03-15",
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, resp.Cards, 1)
	assert.Equal(t, "Patient is 65+", resp.Cards[0].Summary)
}

func TestClient_Call_WithoutPrefetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req cdshooks.CDSRequest
		json.NewDecoder(r.Body).Decode(&req)

		assert.Nil(t, req.Prefetch)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cdshooks.CDSResponse{Cards: []cdshooks.Card{}})
	}))
	defer server.Close()

	ctx := context.Background()
	client := cdshooks.NewClient(server.URL)
	resp, err := client.Call(ctx, "patient-view-age-check",
		cdshooks.PatientViewContext{
			UserID:    "Practitioner/abc",
			PatientID: "Patient/123",
		},
		nil,
	)
	require.NoError(t, err)
	assert.Empty(t, resp.Cards)
}

func TestClient_WithTimeout(t *testing.T) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cdshooks.CDSResponse{Cards: []cdshooks.Card{}})
	}))
	defer slowServer.Close()

	ctx := context.Background()
	client := cdshooks.NewClient(slowServer.URL, cdshooks.WithTimeout(10*time.Millisecond))
	_, err := client.Call(ctx, "test", cdshooks.PatientViewContext{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
}

func TestClient_WithBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cdshooks.CDSResponse{Cards: []cdshooks.Card{}})
	}))
	defer server.Close()

	ctx := context.Background()
	client := cdshooks.NewClient(server.URL, cdshooks.WithBearerToken("test-token"))
	_, err := client.Call(ctx, "test", cdshooks.PatientViewContext{}, nil)
	require.NoError(t, err)
}

func ptr[T any](v T) *T {
	return &v
}
