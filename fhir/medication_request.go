package fhir

type MedicationRequest struct {
	ResourceType              string
	ID                        string
	Status                    string
	Intent                    string
	MedicationCodeableConcept *CodeableConcept
	Subject                   Reference
	AuthoredOn                string
}
