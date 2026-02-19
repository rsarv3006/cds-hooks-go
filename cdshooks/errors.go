package cdshooks

import "fmt"

type ErrUnknownService struct {
	ID string
}

func (e ErrUnknownService) Error() string {
	return fmt.Sprintf("unknown service: %s", e.ID)
}

type ErrMissingPrefetch struct {
	Key string
}

func (e ErrMissingPrefetch) Error() string {
	return fmt.Sprintf("missing required prefetch key: %s", e.Key)
}

type ErrInvalidContext struct {
	Hook  Hook
	Cause error
}

func (e ErrInvalidContext) Error() string {
	return fmt.Sprintf("invalid context for hook %s: %v", e.Hook, e.Cause)
}

func (e ErrInvalidContext) Unwrap() error {
	return e.Cause
}

type ErrInvalidCard struct {
	Field  string
	Reason string
}

func (e ErrInvalidCard) Error() string {
	return fmt.Sprintf("invalid card field %s: %s", e.Field, e.Reason)
}

type ErrFHIRRequest struct {
	URL        string
	StatusCode int
	Body       string
}

func (e ErrFHIRRequest) Error() string {
	return fmt.Sprintf("FHIR request to %s failed with status %d: %s", e.URL, e.StatusCode, e.Body)
}
