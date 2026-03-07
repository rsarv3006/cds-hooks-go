package cdshooks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResponse(t *testing.T) {
	resp := NewResponse()
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.response.Cards)
	assert.Empty(t, resp.response.Cards)
}

func TestResponseBuilder_AddCard(t *testing.T) {
	card := Card{
		Summary:   "Test card",
		Indicator: IndicatorInfo,
		Source:    Source{Label: "Test"},
	}

	resp := NewResponse().AddCard(card).Build()
	assert.Len(t, resp.Cards, 1)
	assert.Equal(t, "Test card", resp.Cards[0].Summary)
}

func TestResponseBuilder_AddMultipleCards(t *testing.T) {
	card1 := Card{Summary: "Card 1", Indicator: IndicatorInfo, Source: Source{Label: "Test"}}
	card2 := Card{Summary: "Card 2", Indicator: IndicatorWarning, Source: Source{Label: "Test"}}

	resp := NewResponse().
		AddCard(card1).
		AddCard(card2).
		Build()

	assert.Len(t, resp.Cards, 2)
	assert.Equal(t, IndicatorInfo, resp.Cards[0].Indicator)
	assert.Equal(t, IndicatorWarning, resp.Cards[1].Indicator)
}

func TestResponseBuilder_AddSystemAction(t *testing.T) {
	action := Action{
		Type:        ActionCreate,
		Description: "Create resource",
	}

	resp := NewResponse().
		AddSystemAction(action).
		Build()

	assert.Len(t, resp.SystemActions, 1)
	assert.Equal(t, ActionCreate, resp.SystemActions[0].Type)
}

func TestResponseBuilder_AddCardAndSystemAction(t *testing.T) {
	card := Card{Summary: "Test", Indicator: IndicatorInfo, Source: Source{Label: "Test"}}
	action := Action{Type: ActionUpdate, Description: "Update"}

	resp := NewResponse().
		AddCard(card).
		AddSystemAction(action).
		Build()

	assert.Len(t, resp.Cards, 1)
	assert.Len(t, resp.SystemActions, 1)
}

func TestResponseBuilder_Build_NilCards(t *testing.T) {
	resp := NewResponse().Build()
	assert.NotNil(t, resp.Cards)
	assert.Empty(t, resp.Cards)
}

func TestEmptyResponse(t *testing.T) {
	resp := EmptyResponse()
	assert.NotNil(t, resp.Cards)
	assert.Empty(t, resp.Cards)
	assert.Empty(t, resp.SystemActions)
}

func TestCDSResponse_JSONMarshal(t *testing.T) {
	resp := CDSResponse{
		Cards: []Card{
			{Summary: "Test", Indicator: IndicatorInfo, Source: Source{Label: "Test"}},
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "cards")
	assert.Contains(t, string(data), "Test")
}

func TestCDSResponse_SystemActionsOmitempty(t *testing.T) {
	resp := CDSResponse{
		Cards: []Card{{Summary: "Test", Indicator: IndicatorInfo, Source: Source{Label: "Test"}}},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.NotContains(t, string(data), "systemActions")
}

func TestCDSResponse_SystemActionsPresent(t *testing.T) {
	resp := CDSResponse{
		Cards: []Card{{Summary: "Test", Indicator: IndicatorInfo, Source: Source{Label: "Test"}}},
		SystemActions: []Action{
			{Type: ActionCreate, Description: "Create"},
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "systemActions")
}

func TestFeedbackOutcome_Constants(t *testing.T) {
	assert.Equal(t, FeedbackOutcome("accepted"), OutcomeAccepted)
	assert.Equal(t, FeedbackOutcome("overridden"), OutcomeOverridden)
}

func TestMarshalFeedbackResponse(t *testing.T) {
	data, err := MarshalFeedbackResponse("ok")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "ok")

	var resp FeedbackResponse
	err = json.Unmarshal(data, &resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
}

func TestFeedbackRequest_JSONMarshal(t *testing.T) {
	req := FeedbackRequest{
		Card:    "card-uuid",
		Outcome: OutcomeAccepted,
		AcceptedSuggestions: []AcceptedSuggestion{
			{ID: "suggestion-uuid"},
		},
		OutcomeTimestamp: "2026-01-01T00:00:00Z",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "card-uuid")
	assert.Contains(t, string(data), "accepted")
}
