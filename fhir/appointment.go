package fhir

type Appointment struct {
	ResourceType    string                   `json:"resourceType"`
	ID              string                   `json:"id,omitempty"`
	Status          string                   `json:"status"`
	ServiceType     []CodeableConcept        `json:"serviceType,omitempty"`
	AppointmentType *CodeableConcept         `json:"appointmentType,omitempty"`
	ReasonCode      []CodeableConcept        `json:"reasonCode,omitempty"`
	Start           string                   `json:"start,omitempty"`
	End             string                   `json:"end,omitempty"`
	MinutesDuration *int                     `json:"minutesDuration,omitempty"`
	Participant     []AppointmentParticipant `json:"participant"`
	Created         string                   `json:"created,omitempty"`
	Comment         string                   `json:"comment,omitempty"`
}

type AppointmentParticipant struct {
	Type     []CodeableConcept `json:"type,omitempty"`
	Actor    *Reference        `json:"actor,omitempty"`
	Required string            `json:"required,omitempty"`
	Status   string            `json:"status"`
}

type Encounter struct {
	ResourceType string                 `json:"resourceType"`
	ID           string                 `json:"id,omitempty"`
	Status       string                 `json:"status"`
	Class        Coding                 `json:"class"`
	Type         []CodeableConcept      `json:"type,omitempty"`
	Subject      *Reference             `json:"subject,omitempty"`
	Participant  []EncounterParticipant `json:"participant,omitempty"`
	Period       *Period                `json:"period,omitempty"`
	ReasonCode   []CodeableConcept      `json:"reasonCode,omitempty"`
}

type EncounterParticipant struct {
	Type       []CodeableConcept `json:"type,omitempty"`
	Period     *Period           `json:"period,omitempty"`
	Individual *Reference        `json:"individual,omitempty"`
}
