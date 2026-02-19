package cdshooks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/your-org/cds-hooks-go/fhir"
)

func TestPrefetch_Get(t *testing.T) {
	raw := map[string]json.RawMessage{
		"patient": json.RawMessage(`{"resourceType":"Patient"}`),
	}
	p := Prefetch{raw: raw}

	val, ok := p.Get("patient")
	assert.True(t, ok)
	assert.Contains(t, string(val), "Patient")

	_, ok = p.Get("nonexistent")
	assert.False(t, ok)
}

func TestPrefetch_Decode(t *testing.T) {
	raw := map[string]json.RawMessage{
		"patient": json.RawMessage(`{"resourceType":"Patient","id":"123"}`),
	}
	p := Prefetch{raw: raw}

	var patient fhir.Patient
	err := p.Decode("patient", &patient)
	assert.NoError(t, err)
	assert.Equal(t, "123", patient.ID)
}

func TestPrefetch_Decode_Missing(t *testing.T) {
	p := Prefetch{raw: make(map[string]json.RawMessage)}

	err := p.Decode("patient", &struct{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestPrefetch_Patient(t *testing.T) {
	raw := map[string]json.RawMessage{
		"patient": json.RawMessage(`{"resourceType":"Patient","id":"123","birthDate":"1990-01-15"}`),
	}
	p := Prefetch{raw: raw}

	patient, err := p.Patient("patient")
	assert.NoError(t, err)
	assert.Equal(t, "123", patient.ID)
}

func TestPrefetch_Bundle(t *testing.T) {
	raw := map[string]json.RawMessage{
		"meds": json.RawMessage(`{"resourceType":"Bundle","type":"searchset","entry":[]}`),
	}
	p := Prefetch{raw: raw}

	bundle, err := p.Bundle("meds")
	assert.NoError(t, err)
	assert.Equal(t, "searchset", bundle.Type)
}

func TestPrefetch_Missing(t *testing.T) {
	raw := map[string]json.RawMessage{
		"patient": json.RawMessage(`{}`),
	}
	p := Prefetch{raw: raw}

	declared := map[string]string{
		"patient": "Patient/{{context.patientId}}",
		"meds":    "MedicationRequest?patient={{context.patientId}}",
	}

	missing := p.Missing(declared)
	assert.Contains(t, missing, "meds")
	assert.NotContains(t, missing, "patient")
}
