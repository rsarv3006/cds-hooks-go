package cdshooks

import (
	"encoding/json"
	"testing"

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
