package cdshooks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrBirthDateEmpty(t *testing.T) {
	assert.Equal(t, "birth date is empty", ErrBirthDateEmpty.Error())
}

func TestErrUnknownService(t *testing.T) {
	err := ErrUnknownService{ID: "test-123"}
	assert.Equal(t, "unknown service: test-123", err.Error())
	assert.Contains(t, err.Error(), "test-123")
}

func TestErrMissingPrefetch(t *testing.T) {
	err := ErrMissingPrefetch{Key: "patient"}
	assert.Equal(t, "missing required prefetch key: patient", err.Error())
	assert.Contains(t, err.Error(), "patient")
}

func TestErrInvalidContext(t *testing.T) {
	innerErr := errors.New("field required")
	err := ErrInvalidContext{
		Hook:  HookPatientView,
		Cause: innerErr,
	}
	assert.Equal(t, "invalid context for hook patient-view: field required", err.Error())
	assert.Contains(t, err.Error(), "patient-view")
	assert.Contains(t, err.Error(), "field required")
}

func TestErrInvalidContext_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := ErrInvalidContext{
		Hook:  HookOrderSign,
		Cause: innerErr,
	}
	assert.Equal(t, innerErr, err.Unwrap())
}

func TestErrInvalidCard(t *testing.T) {
	err := ErrInvalidCard{Field: "Summary", Reason: "required"}
	assert.Equal(t, "invalid card field Summary: required", err.Error())
	assert.Contains(t, err.Error(), "Summary")
	assert.Contains(t, err.Error(), "required")
}

func TestErrFHIRRequest(t *testing.T) {
	err := ErrFHIRRequest{
		URL:        "http://fhir.example.org/Patient/123",
		StatusCode: 404,
		Body:       "Not Found",
	}
	expected := "FHIR request to http://fhir.example.org/Patient/123 failed with status 404: Not Found"
	assert.Equal(t, expected, err.Error())
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Not Found")
}
