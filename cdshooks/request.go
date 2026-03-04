package cdshooks

import (
	"encoding/json"
)

type CDSRequest struct {
	Hook         string          `json:"hook"`
	HookInstance string          `json:"hookInstance"`
	FHIRServer   string          `json:"fhirServer"`
	FHIRAuth     *FHIRAuth       `json:"fhirAuthorization"`
	Context      json.RawMessage `json:"context"`
	Prefetch     Prefetch        `json:"prefetch"`
	Extension    map[string]any  `json:"extension,omitempty"`
}

type FHIRAuth struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	Subject     string `json:"subject"`
	Patient     string `json:"patient,omitempty"`
}

func DecodeContext[T any](ctx json.RawMessage) (T, error) {
	var result T
	err := json.Unmarshal(ctx, &result)
	return result, err
}
