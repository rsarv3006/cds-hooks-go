package cdshooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONRoundtrip_CDSRequest(t *testing.T) {
	jsonData := `{
		"hookInstance": "550e8400-e29b-41d4-a716-446655440000",
		"hook": "patient-view",
		"context": {
			"userId": "Practitioner/123",
			"patientId": "Patient/456"
		},
		"prefetch": {
			"patient": {"resourceType": "Patient", "id": "456"}
		}
	}`

	var req CDSRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", req.HookInstance)
	assert.Equal(t, "patient-view", req.Hook)

	output, err := json.Marshal(req)
	assert.NoError(t, err)

	var req2 CDSRequest
	err = json.Unmarshal(output, &req2)
	assert.NoError(t, err)
	assert.Equal(t, req.HookInstance, req2.HookInstance)
}

func TestJSONRoundtrip_CDSResponse(t *testing.T) {
	resp := CDSResponse{
		Cards: []Card{
			{
				Summary:   "Test card",
				Indicator: "info",
				Source:    Source{Label: "Test"},
			},
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var resp2 CDSResponse
	err = json.Unmarshal(data, &resp2)
	assert.NoError(t, err)
	assert.Len(t, resp2.Cards, 1)
	assert.Equal(t, "Test card", resp2.Cards[0].Summary)
}

func TestJSONRoundtrip_SystemActions_omitempty(t *testing.T) {
	resp := CDSResponse{
		Cards: []Card{},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var resp2 CDSResponse
	err = json.Unmarshal(data, &resp2)
	assert.NoError(t, err)

	assert.NotContains(t, string(data), "systemActions")
}

func TestJSONRoundtrip_SystemActions_Present(t *testing.T) {
	resp := CDSResponse{
		Cards: []Card{},
		SystemActions: []Action{
			{Type: "create", Description: "Create patient"},
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	assert.Contains(t, string(data), "systemActions")

	var resp2 CDSResponse
	err = json.Unmarshal(data, &resp2)
	assert.NoError(t, err)
	assert.Len(t, resp2.SystemActions, 1)
}

func TestJSONRoundtrip_CardWithNilPointers(t *testing.T) {
	card := Card{
		Summary:   "Test",
		Indicator: "info",
		Source:    Source{Label: "Test"},
	}

	data, err := json.Marshal(card)
	assert.NoError(t, err)

	var card2 Card
	err = json.Unmarshal(data, &card2)
	assert.NoError(t, err)
	assert.Equal(t, "Test", card2.Summary)
}

func TestJSONRoundtrip_SuggestionWithNilPointers(t *testing.T) {
	suggestion := Suggestion{
		Label: "Test suggestion",
	}

	data, err := json.Marshal(suggestion)
	assert.NoError(t, err)

	var suggestion2 Suggestion
	err = json.Unmarshal(data, &suggestion2)
	assert.NoError(t, err)
	assert.Equal(t, "Test suggestion", suggestion2.Label)
}

func TestFullLifecycle_PatientView(t *testing.T) {
	title := "Test Service"
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "patient-view-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "A test patient-view service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			card, _ := NewCard("Patient is due for screening", IndicatorInfo).
				WithSource(Source{Label: "CDS Service"}).
				Build()
			return NewResponse().AddCard(card).Build(), nil
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

	var resp CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Cards, 1)
	assert.Equal(t, "Patient is due for screening", resp.Cards[0].Summary)
}

func TestFullLifecycle_OrderSelect(t *testing.T) {
	title := "Order Select Service"
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "order-select-service",
			Hook:        HookOrderSelect,
			Title:       &title,
			Description: "A test order-select service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			card, _ := NewCard("Consider alternative medication", IndicatorWarning).
				WithSource(Source{Label: "CDS Service"}).
				WithSelectionBehavior("at-most-one").
				AddSuggestion(Suggestion{
					Label: "Switch to generic",
					Actions: &[]Action{
						{Type: "create", Description: "Create alternative order"},
					},
				}).
				Build()
			return NewResponse().AddCard(card).Build(), nil
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

	var resp CDSResponse
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
		ServiceEntry{
			Service: Service{
				ID:          "service-one",
				Hook:        HookPatientView,
				Title:       &title1,
				Description: "First service",
			},
			Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
				return EmptyResponse(), nil
			}),
		},
		ServiceEntry{
			Service: Service{
				ID:          "service-two",
				Hook:        HookPatientView,
				Title:       &title2,
				Description: "Second service",
			},
			Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
				card, _ := NewCard("Second service card", IndicatorInfo).
					WithSource(Source{Label: "Service Two"}).
					Build()
				return NewResponse().AddCard(card).Build(), nil
			}),
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string][]Service
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
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "prefetch-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "Service that uses prefetch",
			Prefetch: map[string]string{
				"patient": "Patient/{{context.patientId}}",
			},
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			missing := req.Prefetch.Missing(map[string]string{
				"patient": "Patient/{{context.patientId}}",
			})
			card, _ := NewCard("Missing prefetch: "+strings.Join(missing, ","), IndicatorInfo).
				WithSource(Source{Label: "CDS Service"}).
				Build()
			return NewResponse().AddCard(card).Build(), nil
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

	var resp CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Cards, 1)
	assert.Contains(t, resp.Cards[0].Summary, "patient")
}

func TestFullLifecycle_CardResponseFormat(t *testing.T) {
	title := "Full Card Service"
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "full-card-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "Service returning full card",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			card, err := NewCard("Screening recommendation", IndicatorInfo).
				WithSource(Source{Label: "Guideline Service"}).
				WithSelectionBehavior("at-most-one").
				AddSuggestion(Suggestion{
					Label:         "Order screening",
					IsRecommended: boolPtr(true),
					Actions: &[]Action{
						{Type: ActionCreate, Description: "Create order"},
					},
				}).
				Build()
			if err != nil {
				return EmptyResponse(), err
			}
			return NewResponse().AddCard(card).Build(), nil
		}),
	})

	body := `{"hook":"patient-view","hookInstance":"550e8400-e29b-41d4-a716-446655440000","context":{}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/cds-services/full-card-service",
		strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp CDSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Cards, 1)

	card := resp.Cards[0]
	assert.Equal(t, "Screening recommendation", card.Summary)
	assert.Len(t, *card.Suggestions, 1)
}

func boolPtr(b bool) *bool {
	return &b
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
		ServiceEntry{
			Service: Service{
				ID:          "service-a",
				Hook:        HookPatientView,
				Title:       &title1,
				Description: "Service A description",
			},
			Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
				return EmptyResponse(), nil
			}),
		},
		ServiceEntry{
			Service: Service{
				ID:          "service-b",
				Hook:        HookOrderSelect,
				Title:       &title2,
				Description: "Service B description",
			},
			Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
				return EmptyResponse(), nil
			}),
		},
		ServiceEntry{
			Service: Service{
				ID:          "service-c",
				Hook:        HookOrderSign,
				Title:       &title3,
				Description: "Service C description",
			},
			Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
				return EmptyResponse(), nil
			}),
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	var result map[string][]Service
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result["services"], 3)

	services := result["services"]
	serviceMap := make(map[string]Service)
	for _, s := range services {
		serviceMap[s.ID] = s
	}

	assert.Contains(t, serviceMap, "service-a")
	assert.Contains(t, serviceMap, "service-b")
	assert.Contains(t, serviceMap, "service-c")
	assert.Equal(t, HookPatientView, serviceMap["service-a"].Hook)
}

func TestDiscovery_ServiceWithPrefetch(t *testing.T) {
	title := "Service with Prefetch"
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "prefetch-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "Service with prefetch",
			Prefetch: map[string]string{
				"patient":           "Patient/{{context.patientId}}",
				"medicationRequest": "MedicationRequest?patient={{context.patientId}}",
			},
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/cds-services", nil)
	server.Handler().ServeHTTP(w, r)

	var result map[string][]Service
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
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "test-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			assert.NotNil(t, req.Prefetch)
			missing := req.Prefetch.Missing(map[string]string{"patient": "Patient/123"})
			assert.Contains(t, missing, "patient")
			return EmptyResponse(), nil
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
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "test-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			assert.NotNil(t, req.Prefetch)
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
}

func TestServiceInvocation_UUID(t *testing.T) {
	title := "Test Service"
	var receivedUUID string
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "test-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			receivedUUID = req.HookInstance
			return EmptyResponse(), nil
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

	var result map[string][]Service
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.NotNil(t, result["services"])
	assert.Len(t, result["services"], 0)
}

func TestServiceInvocation_EmptyContext(t *testing.T) {
	title := "Test Service"
	server := NewServer()
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "test-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "A test service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
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
	resp := CDSResponse{
		Cards: []Card{},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(resp)
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), `"cards":[]`)
}

func TestServer_RequestTimeout(t *testing.T) {
	title := "Slow Service"
	server := NewServer(WithRequestTimeout(100))
	server.Register(ServiceEntry{
		Service: Service{
			ID:          "slow-service",
			Hook:        HookPatientView,
			Title:       &title,
			Description: "A slow service",
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			select {
			case <-ctx.Done():
				return EmptyResponse(), ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return EmptyResponse(), nil
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
