package cdshooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	bearerToken string
	fhirServer  string
	fhirAuth    *FHIRAuth
}

type ClientOption func(*Client)

func WithHTTPClient(c *http.Client) ClientOption {
	return func(cl *Client) {
		cl.httpClient = c
	}
}

func WithBearerToken(token string) ClientOption {
	return func(cl *Client) {
		cl.bearerToken = token
	}
}

func WithFHIRServer(baseURL string, auth *FHIRAuth) ClientOption {
	return func(cl *Client) {
		cl.fhirServer = baseURL
		cl.fhirAuth = auth
	}
}

func WithTimeout(d time.Duration) ClientOption {
	return func(cl *Client) {
		cl.httpClient.Timeout = d
	}
}

func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Discover(ctx context.Context) ([]Service, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/cds-services", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery failed with status %d", resp.StatusCode)
	}

	var result struct {
		Services []Service `json:"services"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Services, nil
}

func (c *Client) Call(
	ctx context.Context,
	serviceID string,
	hookCtx any,
	prefetch map[string]any,
) (CDSResponse, error) {
	hookInstance, err := uuid.NewUUID()
	if err != nil {
		return CDSResponse{}, err
	}

	requestBody := map[string]any{
		"hook":         getHookType(hookCtx),
		"hookInstance": hookInstance.String(),
		"context":      hookCtx,
	}

	if c.fhirServer != "" {
		requestBody["fhirServer"] = c.fhirServer
	}

	if c.fhirAuth != nil {
		requestBody["fhirAuthorization"] = c.fhirAuth
	}

	if prefetch != nil {
		requestBody["prefetch"] = prefetch
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return CDSResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/cds-services/"+serviceID, bytes.NewReader(body))
	if err != nil {
		return CDSResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CDSResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CDSResponse{}, fmt.Errorf("service call failed with status %d", resp.StatusCode)
	}

	var result CDSResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return CDSResponse{}, err
	}

	return result, nil
}

func (c *Client) Feedback(
	ctx context.Context,
	serviceID string,
	feedback FeedbackRequest,
) error {
	body, err := json.Marshal(feedback)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/cds-services/"+serviceID+"/feedback", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feedback failed with status %d", resp.StatusCode)
	}

	return nil
}

func getHookType(ctx any) string {
	switch ctx.(type) {
	case PatientViewContext:
		return string(HookPatientView)
	case OrderSelectContext:
		return string(HookOrderSelect)
	case OrderSignContext:
		return string(HookOrderSign)
	case AppointmentBookContext:
		return string(HookAppointmentBook)
	case EncounterStartContext:
		return string(HookEncounterStart)
	case EncounterDischargeContext:
		return string(HookEncounterDischarge)
	default:
		return ""
	}
}
