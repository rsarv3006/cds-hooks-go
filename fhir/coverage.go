package fhir

type Coverage struct {
	ResourceType string          `json:"resourceType"`
	ID           string          `json:"id,omitempty"`
	Status       string          `json:"status"`
	Beneficiary  Reference       `json:"beneficiary"`
	Payor        []Reference     `json:"payor,omitempty"`
	Class        []CoverageClass `json:"class,omitempty"`
	Period       *Period         `json:"period,omitempty"`
}

type CoverageClass struct {
	Type  CodeableConcept `json:"type"`
	Value string          `json:"value"`
	Group string          `json:"group,omitempty"`
	Name  string          `json:"name,omitempty"`
}

type Period struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"`
	Text   string   `json:"text,omitempty"`
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
	if len(c.Coding) > 0 && c.Coding[0].Display != nil {
		return *c.Coding[0].Display
	}
	return c.Text
}

type Coding struct {
	System       string  `json:"system,omitempty"`
	Version      string  `json:"version,omitempty"`
	Code         string  `json:"code"`
	Display      *string `json:"display,omitempty"`
	UserSelected *bool   `json:"userSelected,omitempty"`
}

type Reference struct {
	Reference string `json:"reference,omitempty"`
	Type      string `json:"type,omitempty"`
	Display   string `json:"display,omitempty"`
}

type HumanName struct {
	Use    string   `json:"use,omitempty"`
	Family string   `json:"family,omitempty"`
	Given  []string `json:"given,omitempty"`
	Prefix []string `json:"prefix,omitempty"`
	Suffix []string `json:"suffix,omitempty"`
	Text   string   `json:"text,omitempty"`
}

func (n HumanName) DisplayName() string {
	var family, given string
	if n.Family != "" {
		family = n.Family
	}
	if len(n.Given) > 0 {
		given = n.Given[0]
	}
	if family != "" && given != "" {
		return family + ", " + given
	}
	if n.Text != "" {
		return n.Text
	}
	if given != "" {
		return given
	}
	return family
}

type Identifier struct {
	Use    string `json:"use,omitempty"`
	System string `json:"system,omitempty"`
	Value  string `json:"value"`
}
