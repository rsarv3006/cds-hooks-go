package fhir

type MedicationRequest struct {
	ResourceType              string            `json:"resourceType"`
	ID                        string            `json:"id,omitempty"`
	Status                    string            `json:"status"`
	Intent                    string            `json:"intent"`
	MedicationCodeableConcept *CodeableConcept  `json:"medicationCodeableConcept,omitempty"`
	MedicationReference       *Reference        `json:"medicationReference,omitempty"`
	Subject                   Reference         `json:"subject"`
	Encounter                 *Reference        `json:"encounter,omitempty"`
	AuthoredOn                string            `json:"authoredOn,omitempty"`
	Requester                 *Reference        `json:"requester,omitempty"`
	DoNotPerform              *bool             `json:"doNotPerform,omitempty"`
	ReasonCode                []CodeableConcept `json:"reasonCode,omitempty"`
	SupportingInformation     []Reference       `json:"supportingInformation,omitempty"`
}

type ServiceRequest struct {
	ResourceType          string            `json:"resourceType"`
	ID                    string            `json:"id,omitempty"`
	Status                string            `json:"status"`
	Intent                string            `json:"intent"`
	Code                  *CodeableConcept  `json:"code,omitempty"`
	Subject               Reference         `json:"subject"`
	Encounter             *Reference        `json:"encounter,omitempty"`
	AuthoredOn            string            `json:"authoredOn,omitempty"`
	Requester             *Reference        `json:"requester,omitempty"`
	ReasonCode            []CodeableConcept `json:"reasonCode,omitempty"`
	SupportingInformation []Reference       `json:"supportingInformation,omitempty"`
}

type DiagnosticRequest struct {
	ResourceType string            `json:"resourceType"`
	ID           string            `json:"id,omitempty"`
	Status       string            `json:"status"`
	Intent       string            `json:"intent"`
	Code         *CodeableConcept  `json:"code,omitempty"`
	Subject      Reference         `json:"subject"`
	Encounter    *Reference        `json:"encounter,omitempty"`
	AuthoredOn   string            `json:"authoredOn,omitempty"`
	Requester    *Reference        `json:"requester,omitempty"`
	ReasonCode   []CodeableConcept `json:"reasonCode,omitempty"`
}
