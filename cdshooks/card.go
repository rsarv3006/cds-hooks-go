package cdshooks

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type CardIndicator string

const (
	IndicatorInfo     CardIndicator = "info"
	IndicatorWarning  CardIndicator = "warning"
	IndicatorCritical CardIndicator = "critical"
	IndicatorSuccess  CardIndicator = "success"
)

const (
	ExtCoverageInformation = "http://hl7.org/fhir/us/davinci-crd/StructureDefinition/ext-coverage-information"
	ExtInstructions        = "http://hl7.org/fhir/us/davinci-crd/StructureDefinition/ext-instructions"
	ExtExternalReference   = "http://hl7.org/fhir/us/davinci-crd/StructureDefinition/ext-external-reference"
	ExtProposeAlternate    = "http://hl7.org/fhir/us/davinci-crd/StructureDefinition/ext-propose-alternate"
)

type Card struct {
	UUID              *string         `json:"uuid,omitempty"`
	Summary           string          `json:"summary"`
	Detail            *string         `json:"detail,omitempty"`
	Indicator         CardIndicator   `json:"indicator"`
	Source            Source          `json:"source"`
	Suggestions       *[]Suggestion   `json:"suggestions,omitempty"`
	SelectionBehavior string          `json:"selectionBehavior,omitempty"`
	OverrideReasons   *[]Coding       `json:"overrideReasons,omitempty"`
	Links             *[]Link         `json:"links,omitempty"`
	Extension         []CardExtension `json:"extension,omitempty"`
}

type CardExtension struct {
	URL   string `json:"url"`
	Value any    `json:"value,omitempty"`
}

type Suggestion struct {
	Label         string    `json:"label"`
	UUID          *string   `json:"uuid,omitempty"`
	IsRecommended *bool     `json:"isRecommended,omitempty"`
	Actions       *[]Action `json:"actions,omitempty"`
}

type Action struct {
	Type        ActionType      `json:"type"`
	Description string          `json:"description"`
	Resource    json.RawMessage `json:"resource,omitempty"`
	ResourceID  string          `json:"resourceId,omitempty"`
}

type ActionType string

const (
	ActionCreate ActionType = "create"
	ActionUpdate ActionType = "update"
	ActionDelete ActionType = "delete"
)

type Link struct {
	Label          string   `json:"label"`
	URL            string   `json:"url"`
	Type           LinkType `json:"type,omitempty"`
	AppContext     string   `json:"appContext,omitempty"`
	Autolaunchable *bool    `json:"autolaunchable,omitempty"`
}

type LinkType string

const (
	LinkAbsolute LinkType = "absolute"
	LinkSmart    LinkType = "smart"
)

type Source struct {
	Label string  `json:"label"`
	URL   *string `json:"url,omitempty"`
	Icon  *string `json:"icon,omitempty"`
	Topic *Coding `json:"topic,omitempty"`
}

type OverrideReason struct {
	Reason      *Coding
	UserComment *string
}

type Coding struct {
	System  string  `json:"system,omitempty"`
	Code    string  `json:"code"`
	Display *string `json:"display,omitempty"`
}

type CardBuilder struct {
	card Card
	err  error
}

func NewCard(summary string, indicator CardIndicator) *CardBuilder {
	return &CardBuilder{
		card: Card{
			Summary:   summary,
			Indicator: indicator,
		},
	}
}

func (b *CardBuilder) WithSource(source Source) *CardBuilder {
	if b.err != nil {
		return b
	}
	b.card.Source = source
	return b
}

func (b *CardBuilder) WithDetail(detail string) *CardBuilder {
	if b.err != nil {
		return b
	}
	b.card.Detail = &detail
	return b
}

func (b *CardBuilder) AddSuggestion(suggestion Suggestion) *CardBuilder {
	if b.err != nil {
		return b
	}
	if b.card.Suggestions != nil {
		*b.card.Suggestions = append(*b.card.Suggestions, suggestion)
	}

	b.card.Suggestions = &[]Suggestion{suggestion}
	return b
}

func (b *CardBuilder) AddLink(link Link) *CardBuilder {
	if b.err != nil {
		return b
	}
	if b.card.Links == nil {
		b.card.Links = &[]Link{link}
	} else {
		*b.card.Links = append(*b.card.Links, link)
	}
	return b
}

func (b *CardBuilder) WithSelectionBehavior(behavior string) *CardBuilder {
	if b.err != nil {
		return b
	}
	b.card.SelectionBehavior = behavior
	return b
}

func (b *CardBuilder) AddExtension(url string, value any) *CardBuilder {
	if b.err != nil {
		return b
	}
	b.card.Extension = append(b.card.Extension, CardExtension{
		URL:   url,
		Value: value,
	})
	return b
}

func (b *CardBuilder) Build() (Card, error) {
	if b.err != nil {
		return Card{}, b.err
	}

	if b.card.Summary == "" {
		return Card{}, &ErrInvalidCard{Field: "Summary", Reason: "required"}
	}

	if len(b.card.Summary) > 140 {
		return Card{}, &ErrInvalidCard{Field: "Summary", Reason: "exceeds 140 characters"}
	}

	if b.card.Indicator == "" {
		return Card{}, &ErrInvalidCard{Field: "Indicator", Reason: "required"}
	}

	if b.card.Source.Label == "" {
		return Card{}, &ErrInvalidCard{Field: "Source.Label", Reason: "required"}
	}

	if b.card.Suggestions != nil && len(*b.card.Suggestions) > 0 && b.card.SelectionBehavior == "" {
		return Card{}, &ErrInvalidCard{Field: "SelectionBehavior", Reason: "required when Suggestions present"}
	}

	if b.card.UUID == nil || *b.card.UUID == "" {
		newUUID := uuid.New().String()
		b.card.UUID = &newUUID
	}

	if b.card.Suggestions != nil {
		for i := range *b.card.Suggestions {
			if (*b.card.Suggestions)[i].UUID == nil || *(*b.card.Suggestions)[i].UUID == "" {
				newUUID := uuid.New().String()
				(*b.card.Suggestions)[i].UUID = &newUUID
			}
		}

	}
	return b.card, nil
}

func (b *CardBuilder) MustBuild() Card {
	card, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("MustBuild failed: %v", err))
	}
	return card
}

func NewSuggestion(label string) *SuggestionBuilder {
	return &SuggestionBuilder{
		suggestion: Suggestion{
			Label: label,
		},
	}
}

type SuggestionBuilder struct {
	suggestion Suggestion
	err        error
}

func (b *SuggestionBuilder) WithUUID(uuid string) *SuggestionBuilder {
	if b.err != nil {
		return b
	}
	b.suggestion.UUID = &uuid
	return b
}

func (b *SuggestionBuilder) WithRecommended(recommended bool) *SuggestionBuilder {
	if b.err != nil {
		return b
	}
	b.suggestion.IsRecommended = &recommended
	return b
}

func (b *SuggestionBuilder) AddAction(action Action) *SuggestionBuilder {
	if b.err != nil {
		return b
	}
	if b.suggestion.Actions == nil {
		b.suggestion.Actions = &[]Action{action}
	} else {
		*b.suggestion.Actions = append(*b.suggestion.Actions, action)
	}
	return b
}

func (b *SuggestionBuilder) Build() (Suggestion, error) {
	if b.err != nil {
		return Suggestion{}, b.err
	}

	if b.suggestion.Label == "" {
		return Suggestion{}, &ErrInvalidCard{Field: "Suggestion.Label", Reason: "required"}
	}

	return b.suggestion, nil
}

func NewAction(actionType ActionType, description string) *ActionBuilder {
	return &ActionBuilder{
		action: Action{
			Type:        actionType,
			Description: description,
		},
	}
}

type ActionBuilder struct {
	action Action
	err    error
}

func (b *ActionBuilder) WithResource(resource json.RawMessage) *ActionBuilder {
	if b.err != nil {
		return b
	}
	b.action.Resource = resource
	return b
}

func (b *ActionBuilder) WithResourceID(id string) *ActionBuilder {
	if b.err != nil {
		return b
	}
	b.action.ResourceID = id
	return b
}

func (b *ActionBuilder) Build() Action {
	return b.action
}

func NewDeleteAction(resourceID, description string) Action {
	return Action{
		Type:        ActionDelete,
		ResourceID:  resourceID,
		Description: description,
	}
}
