package cdshooks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookConstants(t *testing.T) {
	assert.Equal(t, Hook("patient-view"), HookPatientView)
	assert.Equal(t, Hook("order-select"), HookOrderSelect)
	assert.Equal(t, Hook("order-sign"), HookOrderSign)
	assert.Equal(t, Hook("appointment-book"), HookAppointmentBook)
	assert.Equal(t, Hook("encounter-start"), HookEncounterStart)
	assert.Equal(t, Hook("encounter-discharge"), HookEncounterDischarge)
}

func TestPatientViewContext_Decode(t *testing.T) {
	jsonData := `{"userId":"Practitioner/123","patientId":"Patient/456","encounterId":"Encounter/789"}`
	var ctx PatientViewContext
	err := json.Unmarshal([]byte(jsonData), &ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
	assert.Equal(t, "Encounter/789", ctx.EncounterID)
}

func TestOrderSelectContext_Decode(t *testing.T) {
	jsonData := `{"userId":"Practitioner/123","patientId":"Patient/456","encounterId":"Encounter/789","selections":["MedicationRequest/1"]}`
	var ctx OrderSelectContext
	err := json.Unmarshal([]byte(jsonData), &ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
	assert.Len(t, ctx.Selections, 1)
	assert.Equal(t, "MedicationRequest/1", ctx.Selections[0])
}

func TestOrderSignContext_Decode(t *testing.T) {
	jsonData := `{"userId":"Practitioner/123","patientId":"Patient/456","encounterId":"Encounter/789"}`
	var ctx OrderSignContext
	err := json.Unmarshal([]byte(jsonData), &ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
}

func TestAppointmentBookContext_Decode(t *testing.T) {
	jsonData := `{"userId":"Practitioner/123","patientId":"Patient/456","encounterId":"Encounter/789"}`
	var ctx AppointmentBookContext
	err := json.Unmarshal([]byte(jsonData), &ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
	assert.Equal(t, "Encounter/789", ctx.EncounterID)
}

func TestEncounterStartContext_Decode(t *testing.T) {
	jsonData := `{"userId":"Practitioner/123","patientId":"Patient/456","encounterId":"Encounter/789"}`
	var ctx EncounterStartContext
	err := json.Unmarshal([]byte(jsonData), &ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
	assert.Equal(t, "Encounter/789", ctx.EncounterID)
}

func TestEncounterDischargeContext_Decode(t *testing.T) {
	jsonData := `{"userId":"Practitioner/123","patientId":"Patient/456","encounterId":"Encounter/789"}`
	var ctx EncounterDischargeContext
	err := json.Unmarshal([]byte(jsonData), &ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
	assert.Equal(t, "Encounter/789", ctx.EncounterID)
}

func TestParseContext_Generic(t *testing.T) {
	jsonData := json.RawMessage(`{"userId":"Practitioner/123","patientId":"Patient/456"}`)

	ctx, err := ParseContext[PatientViewContext](jsonData)
	assert.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
}

func TestParseContext_InvalidJSON(t *testing.T) {
	jsonData := json.RawMessage(`{invalid`)

	_, err := ParseContext[PatientViewContext](jsonData)
	assert.Error(t, err)
}
