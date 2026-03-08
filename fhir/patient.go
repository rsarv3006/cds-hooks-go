package fhir

import (
	"time"
)

type Patient struct {
	ResourceType string         `json:"resourceType"`
	ID           string         `json:"id,omitempty"`
	Active       *bool          `json:"active,omitempty"`
	Name         []HumanName    `json:"name,omitempty"`
	Gender       string         `json:"gender,omitempty"`
	BirthDate    string         `json:"birthDate,omitempty"`
	Deceased     *bool          `json:"deceasedBoolean,omitempty"`
	Identifier   []Identifier   `json:"identifier,omitempty"`
	Telecom      []ContactPoint `json:"telecom,omitempty"`
	Address      []Address      `json:"address,omitempty"`
}

func (p Patient) Age() (int, error) {
	if p.BirthDate == "" {
		return 0, nil
	}

	birth, err := time.Parse("2006-01-02", p.BirthDate)
	if err != nil {
		return 0, err
	}

	today := time.Now()
	age := today.Year() - birth.Year()

	if today.YearDay() < birth.YearDay() {
		age--
	}

	return age, nil
}

func (p Patient) DisplayName() string {
	for _, name := range p.Name {
		if name.Use == "official" || name.Use == "" {
			return name.DisplayName()
		}
	}
	if len(p.Name) > 0 {
		return p.Name[0].DisplayName()
	}
	return ""
}

type ContactPoint struct {
	System string `json:"system,omitempty"`
	Value  string `json:"value,omitempty"`
	Use    string `json:"use,omitempty"`
}

type Address struct {
	Use        string   `json:"use,omitempty"`
	Type       string   `json:"type,omitempty"`
	Line       []string `json:"line,omitempty"`
	City       string   `json:"city,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postalCode,omitempty"`
	Country    string   `json:"country,omitempty"`
}
