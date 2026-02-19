package fhir

import "encoding/json"

type Bundle struct {
	ResourceType string
	ID           string
	Type         string
	Total        *int
	Entry        []BundleEntry
}

type BundleEntry struct {
	FullURL  string
	Resource json.RawMessage
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

func (b Bundle) Count() int {
	return len(b.Entry)
}
