package cdshooks

import (
	"encoding/json"
	"testing"

	fhir "github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "123", *patient.Id)
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
	assert.Equal(t, "123", *patient.Id)
}

func TestPrefetch_Bundle(t *testing.T) {
	raw := map[string]json.RawMessage{
		"meds": json.RawMessage(`{"resourceType":"Bundle","type":"searchset","entry":[]}`),
	}
	p := Prefetch{raw: raw}

	bundle, err := p.Bundle("meds")
	assert.NoError(t, err)
	assert.Equal(t, fhir.BundleTypeSearchset, bundle.Type)
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

func TestPrefetch_UnmarshalFromJSON(t *testing.T) {
	jsonData := `{"hook":"patient-view","hookInstance":"123","context":{},"prefetch":{"patient":{"resourceType":"Patient","id":"123"}}}`
	var req CDSRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	assert.NoError(t, err)

	assert.NotNil(t, req.Prefetch.raw)
	assert.Len(t, req.Prefetch.raw, 1)

	patient, err := req.Prefetch.Patient("patient")
	assert.NoError(t, err)
	assert.Equal(t, "123", *patient.Id)
}

func TestPatientAge(t *testing.T) {
	birthDate := "1990-01-15"
	patient := fhir.Patient{
		BirthDate: &birthDate,
	}

	age, err := PatientAge(patient)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, age, 30)
}

func TestPatientAge_EmptyBirthDate(t *testing.T) {
	patient := fhir.Patient{
		BirthDate: nil,
	}

	_, err := PatientAge(patient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestPatientAge_EmptyStringBirthDate(t *testing.T) {
	emptyDate := ""
	patient := fhir.Patient{
		BirthDate: &emptyDate,
	}

	_, err := PatientAge(patient)
	assert.Error(t, err)
}

func TestPatientAge_InvalidFormat(t *testing.T) {
	invalidDate := "01-15-1990"
	patient := fhir.Patient{
		BirthDate: &invalidDate,
	}

	_, err := PatientAge(patient)
	assert.Error(t, err)
}

func TestBundleEntryCount(t *testing.T) {
	bundle := fhir.Bundle{
		Entry: nil,
	}

	count := BundleEntryCount(bundle)
	assert.Equal(t, 0, count)
}

func TestBundleEntryCount_WithEntries(t *testing.T) {
	bundle := fhir.Bundle{
		Entry: []fhir.BundleEntry{
			{},
			{},
			{},
		},
	}

	count := BundleEntryCount(bundle)
	assert.Equal(t, 3, count)
}

func TestPrefetch_Nil(t *testing.T) {
	var p *Prefetch

	_, ok := p.Get("test")
	assert.False(t, ok)

	err := p.Decode("test", &struct{}{})
	assert.Error(t, err)

	_, err = p.Patient("test")
	assert.Error(t, err)

	_, err = p.Bundle("test")
	assert.Error(t, err)

	missing := p.Missing(map[string]string{"test": "Test"})
	assert.Contains(t, missing, "test")
}
