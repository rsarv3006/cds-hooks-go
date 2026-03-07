package cdshooks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardBuilder_SummaryRequired(t *testing.T) {
	_, err := NewCard("", IndicatorInfo).Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Summary")
}

func TestCardBuilder_SummaryTooLong(t *testing.T) {
	longSummary := string(make([]byte, 141))
	for i := range longSummary {
		longSummary = longSummary[:i] + "a" + longSummary[i+1:]
	}
	_, err := NewCard(longSummary, IndicatorInfo).Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "140 characters")
}

func TestCardBuilder_IndicatorRequired(t *testing.T) {
	_, err := NewCard("Test summary", "").Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Indicator")
}

func TestCardBuilder_SourceLabelRequired(t *testing.T) {
	_, err := NewCard("Test summary", IndicatorInfo).
		WithSource(Source{Label: ""}).
		Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Source.Label")
}

func TestCardBuilder_SelectionBehaviorRequiredWithSuggestions(t *testing.T) {
	_, err := NewCard("Test summary", IndicatorInfo).
		WithSource(Source{Label: "Test"}).
		AddSuggestion(Suggestion{Label: "Test suggestion"}).
		Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SelectionBehavior")
}

func TestCardBuilder_Success(t *testing.T) {
	card, err := NewCard("Test summary", IndicatorInfo).
		WithSource(Source{Label: "Test Source"}).
		WithDetail("Test detail").
		WithSelectionBehavior("at-most-one").
		AddSuggestion(Suggestion{Label: "Test suggestion"}).
		Build()

	assert.NoError(t, err)
	assert.Equal(t, "Test summary", card.Summary)
	assert.Equal(t, IndicatorInfo, card.Indicator)
	assert.NotEmpty(t, card.UUID)
	assert.Len(t, *card.Suggestions, 1)
}

func TestCardBuilder_MustBuildPanics(t *testing.T) {
	assert.Panics(t, func() {
		NewCard("", IndicatorInfo).MustBuild()
	})
}

func TestSuggestionBuilder_LabelRequired(t *testing.T) {
	_, err := NewSuggestion("").Build()
	assert.Error(t, err)
}

func TestSuggestionBuilder_Success(t *testing.T) {
	suggestion, err := NewSuggestion("Test suggestion").
		WithRecommended(true).
		AddAction(Action{Type: ActionCreate, Description: "Create"}).
		Build()

	assert.NoError(t, err)
	assert.Equal(t, "Test suggestion", suggestion.Label)
	assert.NotNil(t, suggestion.IsRecommended)
	assert.True(t, *suggestion.IsRecommended)
	assert.NotNil(t, suggestion.Actions)
	assert.Len(t, *suggestion.Actions, 1)
}

func TestCardBuilder_AddLink(t *testing.T) {
	card, err := NewCard("Test", IndicatorInfo).
		WithSource(Source{Label: "Test"}).
		AddLink(Link{
			Label: "Open App",
			URL:   "https://app.example.com",
			Type:  LinkSmart,
		}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, card.Links)
	assert.Len(t, *card.Links, 1)
	assert.Equal(t, "Open App", (*card.Links)[0].Label)
}

func TestCardBuilder_AddMultipleLinks(t *testing.T) {
	card, err := NewCard("Test", IndicatorInfo).
		WithSource(Source{Label: "Test"}).
		AddLink(Link{Label: "Link 1", URL: "https://1.com", Type: LinkAbsolute}).
		AddLink(Link{Label: "Link 2", URL: "https://2.com", Type: LinkSmart}).
		Build()

	assert.NoError(t, err)
	assert.Len(t, *card.Links, 2)
}

func TestCardBuilder_WithUUID(t *testing.T) {
	card, err := NewCard("Test", IndicatorInfo).
		WithSource(Source{Label: "Test"}).
		Build()
	require.NoError(t, err)

	assert.NotNil(t, card.UUID)
	assert.NotEmpty(t, *card.UUID)
}

func TestCardIndicator_Constants(t *testing.T) {
	assert.Equal(t, CardIndicator("info"), IndicatorInfo)
	assert.Equal(t, CardIndicator("warning"), IndicatorWarning)
	assert.Equal(t, CardIndicator("critical"), IndicatorCritical)
}

func TestLinkType_Constants(t *testing.T) {
	assert.Equal(t, LinkType("absolute"), LinkAbsolute)
	assert.Equal(t, LinkType("smart"), LinkSmart)
}

func TestActionType_Constants(t *testing.T) {
	assert.Equal(t, ActionType("create"), ActionCreate)
	assert.Equal(t, ActionType("update"), ActionUpdate)
	assert.Equal(t, ActionType("delete"), ActionDelete)
}

func TestNewAction(t *testing.T) {
	action := NewAction(ActionUpdate, "Update patient").
		WithResourceID("Patient/123").
		Build()

	assert.Equal(t, ActionUpdate, action.Type)
	assert.Equal(t, "Update patient", action.Description)
	assert.Equal(t, "Patient/123", action.ResourceID)
}

func TestNewAction_WithResource(t *testing.T) {
	resourceJSON := json.RawMessage(`{"resourceType":"Patient"}`)
	action := NewAction(ActionCreate, "Create patient").
		WithResource(resourceJSON).
		Build()

	assert.Equal(t, ActionCreate, action.Type)
	assert.NotEmpty(t, action.Resource)
}

func TestNewDeleteAction(t *testing.T) {
	action := NewDeleteAction("Patient/123", "Remove patient")

	assert.Equal(t, ActionDelete, action.Type)
	assert.Equal(t, "Patient/123", action.ResourceID)
	assert.Equal(t, "Remove patient", action.Description)
}

func TestSource(t *testing.T) {
	url := "https://example.com"
	source := Source{
		Label: "Test Source",
		URL:   &url,
	}

	assert.Equal(t, "Test Source", source.Label)
	assert.NotNil(t, source.URL)
}

func TestCodingStruct(t *testing.T) {
	display := "Test Display"
	coding := Coding{
		System:  "http://example.org",
		Code:    "test",
		Display: &display,
	}

	assert.Equal(t, "http://example.org", coding.System)
	assert.Equal(t, "test", coding.Code)
	assert.NotNil(t, coding.Display)
}

func TestSuggestionBuilder_WithUUID(t *testing.T) {
	suggestion, err := NewSuggestion("Test").
		WithUUID("custom-uuid").
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, suggestion.UUID)
	assert.Equal(t, "custom-uuid", *suggestion.UUID)
}

func TestSuggestionBuilder_AddMultipleActions(t *testing.T) {
	suggestion, err := NewSuggestion("Test").
		AddAction(Action{Type: ActionCreate, Description: "Create 1"}).
		AddAction(Action{Type: ActionUpdate, Description: "Update 1"}).
		Build()

	assert.NoError(t, err)
	assert.Len(t, *suggestion.Actions, 2)
}
