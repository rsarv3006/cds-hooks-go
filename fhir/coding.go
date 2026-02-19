package fhir

type CodeableConcept struct {
	Coding []Coding
	Text   string
}

type Coding struct {
	System  string
	Version string
	Code    string
	Display string
}

type Reference struct {
	Reference string
	Display   string
}

type HumanName struct {
	Use    string
	Family string
	Given  []string
}

func (c CodeableConcept) HasCode(system, code string) bool {
	for _, coding := range c.Coding {
		if coding.System == system && coding.Code == code {
			return true
		}
	}
	return false
}

func (c CodeableConcept) DisplayText() string {
	for _, coding := range c.Coding {
		if coding.Display != "" {
			return coding.Display
		}
	}
	return c.Text
}
