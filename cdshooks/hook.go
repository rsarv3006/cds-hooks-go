package cdshooks

import "encoding/json"

type Hook string

const (
	HookPatientView        Hook = "patient-view"
	HookOrderSelect        Hook = "order-select"
	HookOrderSign          Hook = "order-sign"
	HookAppointmentBook    Hook = "appointment-book"
	HookEncounterStart     Hook = "encounter-start"
	HookEncounterDischarge Hook = "encounter-discharge"
)

type PatientViewContext struct {
	UserID      string
	PatientID   string
	EncounterID string
}

type OrderSelectContext struct {
	UserID      string
	PatientID   string
	EncounterID string
	Selections  []string
	DraftOrders json.RawMessage
}

type OrderSignContext struct {
	UserID      string
	PatientID   string
	EncounterID string
	DraftOrders json.RawMessage
}

type AppointmentBookContext struct {
	UserID       string
	PatientID    string
	EncounterID  string
	Appointments json.RawMessage
}

type EncounterStartContext struct {
	UserID      string
	PatientID   string
	EncounterID string
}

type EncounterDischargeContext struct {
	UserID      string
	PatientID   string
	EncounterID string
}

func ParseContext[T any](raw json.RawMessage) (T, error) {
	var result T
	err := json.Unmarshal(raw, &result)
	return result, err
}
