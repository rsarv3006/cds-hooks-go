package cdshooks

import "encoding/json"

type Hook string

const (
	HookPatientView              Hook = "patient-view"
	HookOrderSelect              Hook = "order-select"
	HookOrderSign                Hook = "order-sign"
	HookOrderReview              Hook = "order-review"
	HookAppointmentBook          Hook = "appointment-book"
	HookEncounterStart           Hook = "encounter-start"
	HookEncounterDischarge       Hook = "encounter-discharge"
	HookMedicationPrescribe      Hook = "medication-prescribe"
	HookMedicationRefill         Hook = "medication-refill"
	HookOrderDispatch            Hook = "order-dispatch"
	HookAllergyIntoleranceCreate Hook = "allergyintolerance-create"
	HookProblemListItemCreate    Hook = "problem-list-item-create"
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
