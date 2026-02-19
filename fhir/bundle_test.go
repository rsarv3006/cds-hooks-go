package fhir

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundle_Resources(t *testing.T) {
	bundle := Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry: []BundleEntry{
			{
				Resource: json.RawMessage(`{"resourceType":"Patient","id":"123"}`),
			},
			{
				Resource: json.RawMessage(`{"resourceType":"Patient","id":"456"}`),
			},
		},
	}

	patients, err := Resources[Patient](bundle)
	assert.NoError(t, err)
	assert.Len(t, patients, 2)
	assert.Equal(t, "123", patients[0].ID)
	assert.Equal(t, "456", patients[1].ID)
}

func TestBundle_Resources_EmptyEntry(t *testing.T) {
	bundle := Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry:        []BundleEntry{},
	}

	patients, err := Resources[Patient](bundle)
	assert.NoError(t, err)
	assert.Len(t, patients, 0)
}

func TestBundle_Resources_NilEntry(t *testing.T) {
	bundle := Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
	}

	patients, err := Resources[Patient](bundle)
	assert.NoError(t, err)
	assert.Len(t, patients, 0)
}

func TestBundle_Resources_InvalidJSON(t *testing.T) {
	bundle := Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Entry: []BundleEntry{
			{
				Resource: json.RawMessage(`{invalid`),
			},
		},
	}

	_, err := Resources[Patient](bundle)
	assert.Error(t, err)
}

func TestBundle_Count(t *testing.T) {
	bundle := Bundle{
		Entry: []BundleEntry{
			{}, {}, {},
		},
	}

	assert.Equal(t, 3, bundle.Count())
}

func TestBundle_Count_Empty(t *testing.T) {
	bundle := Bundle{}
	assert.Equal(t, 0, bundle.Count())
}
