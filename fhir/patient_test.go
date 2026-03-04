package fhir

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatient_Age(t *testing.T) {
	tests := []struct {
		name      string
		birthDate string
		wantAge   int
		wantErr   bool
	}{
		{
			name:      "born 1990-01-01",
			birthDate: "1990-01-01",
			wantAge:   36,
			wantErr:   false,
		},
		{
			name:      "born 2000-12-31",
			birthDate: "2000-12-31",
			wantAge:   25,
			wantErr:   false,
		},
		{
			name:      "empty birth date",
			birthDate: "",
			wantErr:   true,
		},
		{
			name:      "invalid birth date",
			birthDate: "not-a-date",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Patient{BirthDate: tt.birthDate}
			age, err := p.Age()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAge, age)
			}
		})
	}
}

func TestPatient_Age_LeapYear(t *testing.T) {
	p := Patient{BirthDate: "2000-02-29"}
	age, err := p.Age()
	assert.NoError(t, err)
	assert.Equal(t, 26, age)
}

func TestPatient_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		patient  Patient
		expected string
	}{
		{
			name: "official name",
			patient: Patient{
				Name: []HumanName{
					{Use: "official", Family: "Smith", Given: []string{"John"}},
				},
			},
			expected: "Smith, John",
		},
		{
			name: "no official falls back to first",
			patient: Patient{
				Name: []HumanName{
					{Use: "nickname", Family: "Doe", Given: []string{"Jane"}},
				},
			},
			expected: "Doe, Jane",
		},
		{
			name:     "no name",
			patient:  Patient{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.patient.DisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}
