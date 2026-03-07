package cdshooks

import "encoding/json"

type Hook string

const (
	HookPatientView          Hook = "patient-view"
	HookOrderSelect          Hook = "order-select"
	HookOrderSign            Hook = "order-sign"
	HookOrderEdit            Hook = "order-edit"
	HookOrderClose           Hook = "order-close"
	HookAppointmentBook      Hook = "appointment-book"
	HookAppointmentEdit      Hook = "appointment-edit"
	HookEncounterStart       Hook = "encounter-start"
	HookEncounterDischarge   Hook = "encounter-discharge"
	HookMedicationPrescribe  Hook = "medication-prescribe"
	HookMedicationDispense   Hook = "medication-dispense"
	HookMedicationAdminister Hook = "medication-administer"
	HookDiagnosticReport     Hook = "diagnostic-report"
	HookPatientEdit          Hook = "patient-edit"
	HookTask                 Hook = "task"
	HookClaim                Hook = "claim"
	HookSmartConfig          Hook = "smart-config"
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
