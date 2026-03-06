package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cdshooks "github.com/your-org/cds-hooks-go/cdshooks"
	"github.com/your-org/cds-hooks-go/cdshooks/service"
)

func TestService_DiscoveryEndpoint(t *testing.T) {
	svc := newTestService(t)
	server := service.NewServer()
	server.Register(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]cdshooks.Service
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	services := response["services"]
	require.Len(t, services, 1)
	assert.Equal(t, "patient-view-age-check", services[0].ID)
	assert.Equal(t, cdshooks.HookPatientView, services[0].Hook)
}

func TestService_PatientUnder65_NoCards(t *testing.T) {
	svc := newTestService(t)
	server := service.NewServer()
	server.Register(svc)

	body := `{
		"hook": "patient-view",
		"hookInstance": "550e8400-e29b-41d4-a716-446655440000",
		"context": {
			"userId": "Practitioner/1",
			"patientId": "Patient/1"
		},
		"prefetch": {}
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cds-services/patient-view-age-check", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response cdshooks.CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Empty(t, response.Cards)
}

func TestService_PatientOver65_ReturnsCard(t *testing.T) {
	svc := newTestService(t)
	server := service.NewServer()
	server.Register(svc)

	body := `{
		"hook": "patient-view",
		"hookInstance": "550e8400-e29b-41d4-a716-446655440000",
		"context": {
			"userId": "Practitioner/1",
			"patientId": "Patient/1"
		},
		"prefetch": {
			"patient": {"resourceType": "Patient", "id": "1", "birthDate": "1955-03-15"},
			"meds": {"resourceType": "Bundle", "type": "searchset", "entry": []}
		}
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cds-services/patient-view-age-check", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response cdshooks.CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response.Cards, 1)
	assert.Contains(t, response.Cards[0].Summary, "Medication review recommended")
	assert.Equal(t, cdshooks.IndicatorInfo, response.Cards[0].Indicator)
}

func TestService_PatientOver65WithManyMeds_WarningIndicator(t *testing.T) {
	svc := newTestService(t)
	server := service.NewServer()
	server.Register(svc)

	body := `{
		"hook": "patient-view",
		"hookInstance": "550e8400-e29b-41d4-a716-446655440000",
		"context": {
			"userId": "Practitioner/1",
			"patientId": "Patient/1"
		},
		"prefetch": {
			"patient": {"resourceType": "Patient", "id": "1", "birthDate": "1955-03-15"},
			"meds": {"resourceType": "Bundle", "type": "searchset", "entry": [
				{"resource": {}}, {"resource": {}}, {"resource": {}}, {"resource": {}}, {"resource": {}}, {"resource": {}}
			]}
		}
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cds-services/patient-view-age-check", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response cdshooks.CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response.Cards, 1)
	assert.Equal(t, cdshooks.IndicatorWarning, response.Cards[0].Indicator)
}

func TestService_InvalidHookInstance(t *testing.T) {
	svc := newTestService(t)
	server := service.NewServer()
	server.Register(svc)

	body := `{
		"hook": "patient-view",
		"context": {
			"userId": "Practitioner/1",
			"patientId": "Patient/1"
		},
		"prefetch": {}
	}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cds-services/patient-view-age-check", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestService_NotFound(t *testing.T) {
	svc := newTestService(t)
	server := service.NewServer()
	server.Register(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cds-services/nonexistent", nil)
	req.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func newTestService(t *testing.T) cdshooks.ServiceEntry {
	svc, err := cdshooks.NewService("patient-view-age-check").
		ForHook(cdshooks.HookPatientView).
		WithTitle("Patient Age Medication Review").
		WithDescription("Flags patients 65+ for STOPP/START criteria review.").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("meds", "MedicationRequest?subject={{context.patientId}}&status=active").
		HandleFunc(handlePatientView).
		Build()
	require.NoError(t, err)
	return svc
}
