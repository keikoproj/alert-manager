/*
Copyright 2025 Keikoproj authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
