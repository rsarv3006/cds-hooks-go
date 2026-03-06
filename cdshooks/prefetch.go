package cdshooks

import (
	"encoding/json"
	"time"

	fhir "github.com/samply/golang-fhir-models/fhir-models/fhir"
)

type Prefetch struct {
	raw map[string]json.RawMessage
}

func (p *Prefetch) UnmarshalJSON(data []byte) error {
	p.raw = make(map[string]json.RawMessage)
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	return json.Unmarshal(data, &p.raw)
}

func (p *Prefetch) Get(key string) (json.RawMessage, bool) {
	if p == nil || p.raw == nil {
		return nil, false
	}
	val, ok := p.raw[key]
	return val, ok
}

func (p *Prefetch) Decode(key string, target any) error {
	if p == nil || p.raw == nil {
		return &ErrMissingPrefetch{Key: key}
	}
	raw, ok := p.raw[key]
	if !ok {
		return &ErrMissingPrefetch{Key: key}
	}
	return json.Unmarshal(raw, target)
}

func (p *Prefetch) Patient(key string) (fhir.Patient, error) {
	var patient fhir.Patient
	err := p.Decode(key, &patient)
	return patient, err
}

func (p *Prefetch) Bundle(key string) (fhir.Bundle, error) {
	var bundle fhir.Bundle
	err := p.Decode(key, &bundle)
	return bundle, err
}

func (p *Prefetch) Missing(declared map[string]string) []string {
	if p == nil || p.raw == nil {
		missing := make([]string, 0, len(declared))
		for key := range declared {
			missing = append(missing, key)
		}
		return missing
	}
	var missing []string
	for key := range declared {
		if _, ok := p.raw[key]; !ok {
			missing = append(missing, key)
		}
	}
	return missing
}

func PatientAge(patient fhir.Patient) (int, error) {
	if patient.BirthDate == nil || *patient.BirthDate == "" {
		return 0, ErrBirthDateEmpty
	}

	birth, err := time.Parse("2006-01-02", *patient.BirthDate)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	age := now.Year() - birth.Year()

	if now.YearDay() < birth.YearDay() {
		age--
	}

	return age, nil
}

func BundleEntryCount(bundle fhir.Bundle) int {
	if bundle.Entry == nil {
		return 0
	}
	return len(bundle.Entry)
}
