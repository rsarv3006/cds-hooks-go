package cdshooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	svc := NewService("test-service")
	assert.NotNil(t, svc)
	assert.Equal(t, "test-service", svc.service.ID)
	assert.NotNil(t, svc.service.Prefetch)
}

func TestServiceBuilder_ForHook(t *testing.T) {
	svc := NewService("test").ForHook(HookPatientView)
	assert.Equal(t, HookPatientView, svc.service.Hook)
}

func TestServiceBuilder_WithTitle(t *testing.T) {
	svc := NewService("test").WithTitle("My Service")
	assert.NotNil(t, svc.service.Title)
	assert.Equal(t, "My Service", *svc.service.Title)
}

func TestServiceBuilder_WithDescription(t *testing.T) {
	svc := NewService("test").WithDescription("A test service")
	assert.Equal(t, "A test service", svc.service.Description)
}

func TestServiceBuilder_WithPrefetch(t *testing.T) {
	svc := NewService("test").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("meds", "MedicationRequest?patient={{context.patientId}}")

	assert.Equal(t, "Patient/{{context.patientId}}", svc.service.Prefetch["patient"])
	assert.Equal(t, "MedicationRequest?patient={{context.patientId}}", svc.service.Prefetch["meds"])
}

func TestServiceBuilder_WithUsageRequirements(t *testing.T) {
	svc := NewService("test").WithUsageRequirements("Requires FHIR R4")
	assert.Equal(t, "Requires FHIR R4", svc.service.UsageRequirements)
}

func TestServiceBuilder_Handle(t *testing.T) {
	handler := HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
		return EmptyResponse(), nil
	})

	svc := NewService("test").Handle(handler)
	assert.NotNil(t, svc.handler)
}

func TestServiceBuilder_HandleFunc(t *testing.T) {
	svc := NewService("test").HandleFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
		return EmptyResponse(), nil
	})
	assert.NotNil(t, svc.handler)
}

func TestServiceBuilder_Build_Success(t *testing.T) {
	title := "Test Service"
	entry, err := NewService("test-service").
		ForHook(HookPatientView).
		WithTitle(title).
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		HandleFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}).
		Build()

	require.NoError(t, err)
	assert.Equal(t, "test-service", entry.Service.ID)
	assert.Equal(t, HookPatientView, entry.Service.Hook)
	assert.Equal(t, title, *entry.Service.Title)
	assert.NotNil(t, entry.Handler)
}

func TestServiceBuilder_Build_MissingID(t *testing.T) {
	_, err := NewService("").
		ForHook(HookPatientView).
		WithTitle("Test").
		HandleFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}).
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Service.ID")
}

func TestServiceBuilder_Build_MissingHook(t *testing.T) {
	_, err := NewService("test").
		WithTitle("Test").
		HandleFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}).
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Service.Hook")
}

func TestServiceBuilder_Build_MissingTitle(t *testing.T) {
	_, err := NewService("test").
		ForHook(HookPatientView).
		HandleFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}).
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Service.Title")
}

func TestServiceBuilder_Build_MissingHandler(t *testing.T) {
	_, err := NewService("test").
		ForHook(HookPatientView).
		WithTitle("Test").
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Service.Handler")
}

func TestServiceBuilder_ChainedCalls(t *testing.T) {
	title := "Chained Service"
	entry, err := NewService("chained").
		ForHook(HookOrderSelect).
		WithTitle(title).
		WithDescription("Description").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithUsageRequirements("Requires patient data").
		HandleFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}).
		Build()

	require.NoError(t, err)
	assert.Equal(t, "chained", entry.Service.ID)
	assert.Equal(t, HookOrderSelect, entry.Service.Hook)
	assert.Equal(t, title, *entry.Service.Title)
	assert.Equal(t, "Description", entry.Service.Description)
	assert.Equal(t, "Patient/{{context.patientId}}", entry.Service.Prefetch["patient"])
	assert.Equal(t, "Requires patient data", entry.Service.UsageRequirements)
}

func TestHandlerFunc_Handle(t *testing.T) {
	handlerCalled := false
	handler := HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
		handlerCalled = true
		return NewResponse().AddCard(Card{
			Summary:   "Response card",
			Indicator: IndicatorInfo,
			Source:    Source{Label: "Test"},
		}).Build(), nil
	})

	resp, err := handler.Handle(context.Background(), CDSRequest{})
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Len(t, resp.Cards, 1)
}

func TestServiceEntry(t *testing.T) {
	entry := ServiceEntry{
		Service: Service{
			ID:   "test",
			Hook: HookPatientView,
		},
		Handler: HandlerFunc(func(ctx context.Context, req CDSRequest) (CDSResponse, error) {
			return EmptyResponse(), nil
		}),
	}

	assert.Equal(t, "test", entry.Service.ID)
	assert.NotNil(t, entry.Handler)
}
