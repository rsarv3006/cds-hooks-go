package cdshooks

import "encoding/json"

type CDSResponse struct {
	Cards         []Card   `json:"cards"`
	SystemActions []Action `json:"systemActions,omitempty"`
}

type ResponseBuilder struct {
	response CDSResponse
}

func NewResponse() *ResponseBuilder {
	return &ResponseBuilder{
		response: CDSResponse{
			Cards: make([]Card, 0),
		},
	}
}

func (b *ResponseBuilder) AddCard(card Card) *ResponseBuilder {
	b.response.Cards = append(b.response.Cards, card)
	return b
}

func (b *ResponseBuilder) AddSystemAction(action Action) *ResponseBuilder {
	b.response.SystemActions = append(b.response.SystemActions, action)
	return b
}

func (b *ResponseBuilder) Build() CDSResponse {
	if b.response.Cards == nil {
		b.response.Cards = []Card{}
	}
	return b.response
}

func EmptyResponse() CDSResponse {
	return CDSResponse{
		Cards: []Card{},
	}
}

type FeedbackRequest struct {
	Card                string               `json:"card"`
	Outcome             FeedbackOutcome      `json:"outcome"`
	AcceptedSuggestions []AcceptedSuggestion `json:"acceptedSuggestions"`
	OverrideReason      *OverrideReason      `json:"overrideReason"`
	OutcomeTimestamp    string               `json:"outcomeTimestamp"`
}

type FeedbackOutcome string

const (
	OutcomeAccepted   FeedbackOutcome = "accepted"
	OutcomeOverridden FeedbackOutcome = "overridden"
)

type AcceptedSuggestion struct {
	ID string `json:"id"`
}

type FeedbackResponse struct {
	Status string `json:"status"`
}

func MarshalFeedbackResponse(status string) ([]byte, error) {
	return json.Marshal(FeedbackResponse{Status: status})
}
