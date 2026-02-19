package fhir

import (
	"errors"
	"time"
)

var ErrBirthDateEmpty = errors.New("birth date is empty")

type Patient struct {
	ResourceType string
	ID           string
	Active       *bool
	Name         []HumanName
	Gender       string
	BirthDate    string
	Deceased     *bool
}

func (p Patient) Age() (int, error) {
	if p.BirthDate == "" {
		return 0, ErrBirthDateEmpty
	}

	birth, err := time.Parse("2006-01-02", p.BirthDate)
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

func (p Patient) DisplayName() string {
	for _, name := range p.Name {
		if name.Use == "official" || name.Use == "" {
			return formatName(name)
		}
	}

	if len(p.Name) > 0 {
		return formatName(p.Name[0])
	}

	return ""
}

func formatName(name HumanName) string {
	var result string
	if name.Family != "" {
		result = name.Family
	}
	if len(name.Given) > 0 {
		if result != "" {
			result += ", "
		}
		result += name.Given[0]
	}
	return result
}
