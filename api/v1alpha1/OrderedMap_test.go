package v1alpha1_test

import (
	"encoding/json"
	"testing"

	"github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestOrderedMap_MarshalJSON(t *testing.T) {
	// Create an OrderedMap with some key-value pairs
	om := v1alpha1.OrderedMap{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// Marshal the OrderedMap to JSON
	data, err := json.Marshal(om)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	// The data should be a valid JSON object
	var unmarshalled map[string]string
	err = json.Unmarshal(data, &unmarshalled)
	assert.NoError(t, err)

	// The unmarshalled data should contain all the original keys and values
	assert.Equal(t, "value1", unmarshalled["key1"])
	assert.Equal(t, "value2", unmarshalled["key2"])
	assert.Equal(t, "value3", unmarshalled["key3"])
}

func TestOrderedMap_UnmarshalJSON(t *testing.T) {
	// Create a JSON string representing an OrderedMap
	jsonStr := `{"key1":"value1","key2":"value2","key3":"value3"}`

	// Unmarshal the JSON into an OrderedMap
	var om v1alpha1.OrderedMap
	err := json.Unmarshal([]byte(jsonStr), &om)
	assert.NoError(t, err)

	// Verify the OrderedMap contains the expected keys and values
	assert.Equal(t, "value1", om["key1"])
	assert.Equal(t, "value2", om["key2"])
	assert.Equal(t, "value3", om["key3"])
}

func TestOrderedMap_UnmarshalJSON_Error(t *testing.T) {
	// Create an invalid JSON string
	jsonStr := `{"key1":value1"}`

	// Unmarshal the JSON into an OrderedMap
	var om v1alpha1.OrderedMap
	err := json.Unmarshal([]byte(jsonStr), &om)
	assert.Error(t, err)
}

func TestOrderedMap_RoundTrip(t *testing.T) {
	// Create an OrderedMap with some key-value pairs
	original := v1alpha1.OrderedMap{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// Marshal the OrderedMap to JSON
	data, err := json.Marshal(original)
	assert.NoError(t, err)

	// Unmarshal the JSON back into a new OrderedMap
	var unmarshalled v1alpha1.OrderedMap
	err = json.Unmarshal(data, &unmarshalled)
	assert.NoError(t, err)

	// The unmarshalled OrderedMap should contain all the original keys and values
	for k, v := range original {
		assert.Equal(t, v, unmarshalled[k])
	}
}
