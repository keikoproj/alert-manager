package v1alpha1

import (
	"encoding/json"
	"github.com/emirpasic/gods/maps/treemap"
)

// OrderedMap is a TreeMap implementation of map with string comparision
// +kubebuilder:object:generate=false
type OrderedMap map[string]string

// MarshalJSON function is a custom implementation of json.Marshal for OrderedMap
func (s OrderedMap) MarshalJSON() ([]byte, error) {
	m := treemap.NewWithStringComparator()

	for k, v := range s {
		m.Put(k, v)
	}

	return m.ToJSON()
}

// UnmarshalJson function is a custom implementation of json to unmarshal OrderedMap
func (s *OrderedMap) UnmarshalJSON(b []byte) error {
	// Just regular unmarshal into map[string]string should work fine
	var foo map[string]string
	if err := json.Unmarshal(b, &foo); err != nil {
		return err
	}
	*s = foo
	return nil
}
