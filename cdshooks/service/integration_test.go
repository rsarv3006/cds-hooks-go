package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cdshooks "github.com/your-org/cds-hooks-go/cdshooks"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestFullLifecycle_PatientView(t *testing.T) {
	title := "Test Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "patient-view-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test patient-view service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			card, _ := cdshooks.NewCard("Patient is due for screening", cdshooks.IndicatorInfo).
				WithSource(cdshooks.Source{Label: "CDS Service"}).
				Build()
			return cdshooks.NewResponse().AddCard(card).Build(), nil
		}),
	})

	body := `{
		"hook": "patient-view",
		"hookInstance": "550e8400-e29b-41d4-a716-446655440000",
		"context": {
			"userId": "Practitioner/123",
			"patientId": "Patient/456",
			"encounterId": "Encounter/789"
		}
	}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/patient-view-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp cdshooks.CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Cards, 1)
	assert.Equal(t, "Patient is due for screening", resp.Cards[0].Summary)
}

func TestFullLifecycle_OrderSelect(t *testing.T) {
	title := "Order Select Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "order-select-service",
			Hook:        cdshooks.HookOrderSelect,
			Title:       &title,
			Description: "A test order-select service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			card, _ := cdshooks.NewCard("Consider alternative medication", cdshooks.IndicatorWarning).
				WithSource(cdshooks.Source{Label: "CDS Service"}).
				WithSelectionBehavior("at-most-one").
				AddSuggestion(cdshooks.Suggestion{
					Label: "Switch to generic",
					Actions: &[]cdshooks.Action{
						{Type: cdshooks.ActionCreate, Description: "Create alternative order"},
					},
				}).
				Build()
			return cdshooks.NewResponse().AddCard(card).Build(), nil
		}),
	})

	body := `{
		"hook": "order-select",
		"hookInstance": "550e8400-e29b-41d4-a716-446655440001",
		"context": {
			"userId": "Practitioner/123",
			"patientId": "Patient/456",
			"selections": ["MedicationRequest/1"]
		}
	}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/order-select-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp cdshooks.CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Cards, 1)
	assert.NotNil(t, resp.Cards[0].Suggestions)
}

func TestFullLifecycle_MultipleServices(t *testing.T) {
	title1 := "Service One"
	title2 := "Service Two"

	server := NewServer()
	server.Register(
		cdshooks.ServiceEntry{
			Service: cdshooks.Service{
				ID:          "service-one",
				Hook:        cdshooks.HookPatientView,
				Title:       &title1,
				Description: "First service",
			},
			Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
				return cdshooks.EmptyResponse(), nil
			}),
		},
		cdshooks.ServiceEntry{
			Service: cdshooks.Service{
				ID:          "service-two",
				Hook:        cdshooks.HookPatientView,
				Title:       &title2,
				Description: "Second service",
			},
			Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
				card, _ := cdshooks.NewCard("Second service card", cdshooks.IndicatorInfo).
					WithSource(cdshooks.Source{Label: "Service Two"}).
					Build()
				return cdshooks.NewResponse().AddCard(card).Build(), nil
			}),
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string][]cdshooks.Service
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result["services"], 2)

	w2 := httptest.NewRecorder()
	body := `{"hook":"patient-view","hookInstance":"550e8400-e29b-41d4-a716-446655440000","context":{}}`
	r2 := httptest.NewRequest("POST", "/cds-services/service-two",
		strings.NewReader(body))
	r2.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w2, r2)

	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestFullLifecycle_WithPrefetch(t *testing.T) {
	title := "Prefetch Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "prefetch-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "Service that uses prefetch",
			Prefetch: map[string]string{
				"patient": "Patient/{{context.patientId}}",
			},
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			missing := req.Prefetch.Missing(map[string]string{
				"patient": "Patient/{{context.patientId}}",
			})
			card, _ := cdshooks.NewCard("Missing prefetch: "+strings.Join(missing, ","), cdshooks.IndicatorInfo).
				WithSource(cdshooks.Source{Label: "CDS Service"}).
				Build()
			return cdshooks.NewResponse().AddCard(card).Build(), nil
		}),
	})

	body := `{
		"hook": "patient-view",
		"hookInstance": "550e8400-e29b-41d4-a716-446655440000",
		"context": {
			"patientId": "Patient/456"
		},
		"prefetch": {
			"patient": {"resourceType": "Patient", "id": "456", "birthDate": "1990-01-01"}
		}
	}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/prefetch-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp cdshooks.CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Cards, 1)
	assert.Contains(t, resp.Cards[0].Summary, "patient")
}

func TestFullLifecycle_CardResponseFormat(t *testing.T) {
	title := "Full Card Service"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "full-card-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "Service returning full card",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			card, err := cdshooks.NewCard("Screening recommendation", cdshooks.IndicatorInfo).
				WithSource(cdshooks.Source{Label: "Guideline Service"}).
				WithSelectionBehavior("at-most-one").
				AddSuggestion(cdshooks.Suggestion{
					Label:         "Order screening",
					IsRecommended: boolPtr(true),
					Actions: &[]cdshooks.Action{
						{Type: cdshooks.ActionCreate, Description: "Create order"},
					},
				}).
				Build()
			if err != nil {
				return cdshooks.EmptyResponse(), err
			}
			return cdshooks.NewResponse().AddCard(card).Build(), nil
		}),
	})

	body := `{"hook":"patient-view","hookInstance":"550e8400-e29b-41d4-a716-446655440000","context":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/full-card-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp cdshooks.CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Cards, 1)

	card := resp.Cards[0]
	assert.Equal(t, "Screening recommendation", card.Summary)
	assert.Len(t, *card.Suggestions, 1)
}

func TestCORS_SpecificOrigin(t *testing.T) {
	server := NewServer(WithCORSOrigins("http://example.com", "http://test.com"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/cds-services", nil)
	r.Header.Set("Origin", "http://example.com")
	r.Header.Set("Access-Control-Request-Method", "POST")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	server := NewServer(WithCORSOrigins("http://allowed.com"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/cds-services", nil)
	r.Header.Set("Origin", "http://disallowed.com")
	r.Header.Set("Access-Control-Request-Method", "POST")
	server.Handler().ServeHTTP(w, r)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_NoOrigin(t *testing.T) {
	server := NewServer(WithCORSOrigins("http://example.com"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/cds-services", nil)
	r.Header.Set("Access-Control-Request-Method", "POST")
	server.Handler().ServeHTTP(w, r)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_PreflightRequest(t *testing.T) {
	server := NewServer(WithCORSOrigins("*"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/cds-services/test-service", nil)
	r.Header.Set("Origin", "http://example.com")
	r.Header.Set("Access-Control-Request-Method", "POST")
	r.Header.Set("Access-Control-Request-Headers", "Content-Type")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Content-Type, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET, POST, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDiscovery_MultipleServices(t *testing.T) {
	title1 := "Service A"
	title2 := "Service B"
	title3 := "Service C"

	server := NewServer()
	server.Register(
		cdshooks.ServiceEntry{
			Service: cdshooks.Service{
				ID:          "service-a",
				Hook:        cdshooks.HookPatientView,
				Title:       &title1,
				Description: "Service A description",
			},
			Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
				return cdshooks.EmptyResponse(), nil
			}),
		},
		cdshooks.ServiceEntry{
			Service: cdshooks.Service{
				ID:          "service-b",
				Hook:        cdshooks.HookOrderSelect,
				Title:       &title2,
				Description: "Service B description",
			},
			Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
				return cdshooks.EmptyResponse(), nil
			}),
		},
		cdshooks.ServiceEntry{
			Service: cdshooks.Service{
				ID:          "service-c",
				Hook:        cdshooks.HookOrderSign,
				Title:       &title3,
				Description: "Service C description",
			},
			Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
				return cdshooks.EmptyResponse(), nil
			}),
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	var result map[string][]cdshooks.Service
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result["services"], 3)

	services := result["services"]
	serviceMap := make(map[string]cdshooks.Service)
	for _, s := range services {
		serviceMap[s.ID] = s
	}

	assert.Contains(t, serviceMap, "service-a")
	assert.Contains(t, serviceMap, "service-b")
	assert.Contains(t, serviceMap, "service-c")
	assert.Equal(t, cdshooks.HookPatientView, serviceMap["service-a"].Hook)
}

func TestDiscovery_ServiceWithPrefetch(t *testing.T) {
	title := "Service with Prefetch"
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "prefetch-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "Service with prefetch",
			Prefetch: map[string]string{
				"patient":           "Patient/{{context.patientId}}",
				"medicationRequest": "MedicationRequest?patient={{context.patientId}}",
			},
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			return cdshooks.EmptyResponse(), nil
		}),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	var result map[string][]cdshooks.Service
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result["services"], 1)

	service := result["services"][0]
	assert.NotNil(t, service.Prefetch)
	assert.Equal(t, "Patient/{{context.patientId}}", service.Prefetch["patient"])
}

func TestRequestBody_EmptyPrefetch(t *testing.T) {
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
			assert.NotNil(t, req.Prefetch)
			missing := req.Prefetch.Missing(map[string]string{"patient": "Patient/123"})
			assert.Contains(t, missing, "patient")
			return cdshooks.EmptyResponse(), nil
		}),
	})

	body := `{"hook":"patient-view","hookInstance":"550e8400-e29b-41d4-a716-446655440000","context":{},"prefetch":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestBody_NoPrefetch(t *testing.T) {
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
			assert.NotNil(t, req.Prefetch)
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
}

func TestServiceInvocation_UUID(t *testing.T) {
	title := "Test Service"
	var receivedUUID string
	server := NewServer()
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "test-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			receivedUUID = req.HookInstance
			return cdshooks.EmptyResponse(), nil
		}),
	})

	testUUID := "550e8400-e29b-41d4-a716-446655440000"
	body := `{"hook":"patient-view","hookInstance":"` + testUUID + `","context":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testUUID, receivedUUID)
}

func TestDiscovery_EmptyRegistry(t *testing.T) {
	server := NewServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string][]cdshooks.Service
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.NotNil(t, result["services"])
	assert.Len(t, result["services"], 0)
}

func TestServiceInvocation_EmptyContext(t *testing.T) {
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

	body := `{"hook":"patient-view","hookInstance":"550e8400-e29b-41d4-a716-446655440000"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/test-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJSON_EncodeResponse(t *testing.T) {
	resp := cdshooks.CDSResponse{
		Cards: []cdshooks.Card{},
	}

	_, err := json.Marshal(resp)
	assert.NoError(t, err)
}

func TestServer_RequestTimeout(t *testing.T) {
	title := "Slow Service"
	server := NewServer(WithRequestTimeout(100))
	server.Register(cdshooks.ServiceEntry{
		Service: cdshooks.Service{
			ID:          "slow-service",
			Hook:        cdshooks.HookPatientView,
			Title:       &title,
			Description: "A slow service",
		},
		Handler: cdshooks.HandlerFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
			select {
			case <-ctx.Done():
				return cdshooks.EmptyResponse(), ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return cdshooks.EmptyResponse(), nil
			}
		}),
	})

	body := `{"hook":"patient-view","hookInstance":"550e8400-e29b-41d4-a716-446655440000","context":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/slow-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
