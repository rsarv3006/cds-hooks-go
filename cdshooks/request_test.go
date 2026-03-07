package cdshooks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeContext(t *testing.T) {
	jsonData := json.RawMessage(`{"userId":"Practitioner/123","patientId":"Patient/456"}`)
	ctx, err := DecodeContext[PatientViewContext](jsonData)

	require.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
}

func TestDecodeContext_InvalidJSON(t *testing.T) {
	jsonData := json.RawMessage(`{invalid}`)
	_, err := DecodeContext[PatientViewContext](jsonData)

	assert.Error(t, err)
}

func TestDecodeContext_Empty(t *testing.T) {
	jsonData := json.RawMessage(`{}`)
	ctx, err := DecodeContext[PatientViewContext](jsonData)

	require.NoError(t, err)
	assert.Empty(t, ctx.UserID)
}

func TestParseContext(t *testing.T) {
	jsonData := json.RawMessage(`{"userId":"Practitioner/123","patientId":"Patient/456"}`)
	ctx, err := ParseContext[PatientViewContext](jsonData)

	require.NoError(t, err)
	assert.Equal(t, "Practitioner/123", ctx.UserID)
	assert.Equal(t, "Patient/456", ctx.PatientID)
}

func TestCDSRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"hook": "patient-view",
		"hookInstance": "550e8400-e29b-41d4-a716-446655440000",
		"context": {"userId": "Practitioner/123", "patientId": "Patient/456"}
	}`

	var req CDSRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	require.NoError(t, err)
	assert.Equal(t, "patient-view", req.Hook)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", req.HookInstance)
}

func TestCDSRequest_WithPrefetch(t *testing.T) {
	jsonData := `{
		"hook": "patient-view",
		"hookInstance": "123",
		"context": {},
		"prefetch": {"patient": {"resourceType": "Patient", "id": "123"}}
	}`

	var req CDSRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	require.NoError(t, err)
	assert.NotNil(t, req.Prefetch)
}

func TestCDSRequest_NoPrefetch(t *testing.T) {
	jsonData := `{
		"hook": "patient-view",
		"hookInstance": "123",
		"context": {}
	}`

	var req CDSRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	require.NoError(t, err)
	assert.Nil(t, req.Prefetch)
}

func TestFHIRAuth(t *testing.T) {
	jsonData := `{
		"access_token": "token123",
		"token_type": "Bearer",
		"expires_in": 300,
		"scope": "user/Patient.read",
		"subject": "cds-service"
	}`

	var auth FHIRAuth
	err := json.Unmarshal([]byte(jsonData), &auth)

	require.NoError(t, err)
	assert.Equal(t, "token123", auth.AccessToken)
	assert.Equal(t, "Bearer", auth.TokenType)
	assert.Equal(t, 300, auth.ExpiresIn)
	assert.Equal(t, "user/Patient.read", auth.Scope)
	assert.Equal(t, "cds-service", auth.Subject)
}
