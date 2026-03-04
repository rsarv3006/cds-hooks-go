package cdshooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
