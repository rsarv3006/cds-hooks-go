package fhir

type AllergyIntolerance struct {
	ResourceType   string
	ID             string
	ClinicalStatus CodeableConcept
	Code           CodeableConcept
	Patient        Reference
	Reaction       []AllergyReaction
}

type AllergyReaction struct {
	Substance     *CodeableConcept
	Manifestation []CodeableConcept
	Severity      string
}
