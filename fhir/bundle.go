package fhir

import (
	"encoding/json"
)

type Bundle struct {
	ResourceType string        `json:"resourceType"`
	ID           string        `json:"id,omitempty"`
	Type         string        `json:"type"`
	Total        *int          `json:"total,omitempty"`
	Link         []BundleLink  `json:"link,omitempty"`
	Entry        []BundleEntry `json:"entry,omitempty"`
}

type BundleLink struct {
	Relation string `json:"relation"`
	URL      string `json:"url"`
}

type BundleEntry struct {
	FullURL  string          `json:"fullUrl,omitempty"`
	Resource json.RawMessage `json:"resource,omitempty"`
	Search   *BundleSearch   `json:"search,omitempty"`
}

type BundleSearch struct {
	Mode string `json:"mode,omitempty"`
}

func (b Bundle) Count() int {
	if b.Total != nil {
		return *b.Total
	}
	return len(b.Entry)
}

func Resources[T any](b Bundle) ([]T, error) {
	result := make([]T, 0, len(b.Entry))
	for _, entry := range b.Entry {
		if entry.Resource == nil {
			continue
		}
		var resource T
		if err := json.Unmarshal(entry.Resource, &resource); err != nil {
			return nil, err
		}
		result = append(result, resource)
	}
	return result, nil
}
