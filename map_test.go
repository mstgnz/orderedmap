package orderedmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

func TestOrderedMap_BasicOperations(t *testing.T) {
	om := NewOrderedMap()

	// Test Set and Get
	t.Run("Set and Get", func(t *testing.T) {
		om.Set("key1", "value1")
		if val, exists := om.Get("key1"); !exists || val != "value1" {
			t.Errorf("Expected value1, got %v", val)
		}
	})

	// Test non-existent key
	t.Run("Get Non-existent Key", func(t *testing.T) {
		if val, exists := om.Get("nonexistent"); exists || val != nil {
			t.Errorf("Expected nil and false for non-existent key")
		}
	})

	// Test update existing key
	t.Run("Update Existing Key", func(t *testing.T) {
		om.Set("key1", "updated_value")
		if val, exists := om.Get("key1"); !exists || val != "updated_value" {
			t.Errorf("Expected updated_value, got %v", val)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		om.Delete("key1")
		if val, exists := om.Get("key1"); exists || val != nil {
			t.Errorf("Expected key to be deleted")
		}
	})
}

func TestOrderedMap_Order(t *testing.T) {
	om := NewOrderedMap()

	// Add elements in specific order
	elements := []struct {
		key   string
		value int
	}{
		{"first", 1},
		{"second", 2},
		{"third", 3},
	}

	for _, elem := range elements {
		om.Set(elem.key, elem.value)
	}

	// Test Keys order
	t.Run("Keys Order", func(t *testing.T) {
		keys := om.Keys()
		if len(keys) != len(elements) {
			t.Errorf("Expected %d keys, got %d", len(elements), len(keys))
		}
		for i, elem := range elements {
			if keys[i] != elem.key {
				t.Errorf("Expected key %s at position %d, got %v", elem.key, i, keys[i])
			}
		}
	})

	// Test Values order
	t.Run("Values Order", func(t *testing.T) {
		values := om.Values()
		if len(values) != len(elements) {
			t.Errorf("Expected %d values, got %d", len(elements), len(values))
		}
		for i, elem := range elements {
			if values[i] != elem.value {
				t.Errorf("Expected value %d at position %d, got %v", elem.value, i, values[i])
			}
		}
	})
}

func TestOrderedMap_ConcurrentOperations(t *testing.T) {
	om := NewOrderedMap()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent writes
	t.Run("Concurrent Writes", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				key := val
				om.Set(key, val)
			}(i)
		}
		wg.Wait()

		if om.Len() != numGoroutines {
			t.Errorf("Expected length %d, got %d", numGoroutines, om.Len())
		}
	})

	// Test concurrent reads
	t.Run("Concurrent Reads", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				key := val
				if _, exists := om.Get(key); !exists {
					t.Errorf("Key %v should exist", key)
				}
			}(i)
		}
		wg.Wait()
	})

	// Test concurrent reads and writes
	t.Run("Concurrent Reads and Writes", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(2)
			// Reader
			go func(val int) {
				defer wg.Done()
				om.Get(val)
			}(i)
			// Writer
			go func(val int) {
				defer wg.Done()
				om.Set(val, val*2)
			}(i)
		}
		wg.Wait()
	})

	// Test concurrent deletes
	t.Run("Concurrent Deletes", func(t *testing.T) {
		initialLen := om.Len()
		for i := 0; i < numGoroutines/2; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				om.Delete(val)
			}(i)
		}
		wg.Wait()

		expectedLen := initialLen - numGoroutines/2
		if om.Len() != expectedLen {
			t.Errorf("Expected length %d after deletes, got %d", expectedLen, om.Len())
		}
	})
}

func TestOrderedMap_String(t *testing.T) {
	om := NewOrderedMap()
	om.Set("key1", 1)
	om.Set("key2", 2)

	str := om.String()
	expected := "{key1: 1, key2: 2}"
	if str != expected {
		t.Errorf("Expected string representation %s, got %s", expected, str)
	}
}

func TestOrderedMap_EmptyOperations(t *testing.T) {
	om := NewOrderedMap()

	t.Run("Empty Map Operations", func(t *testing.T) {
		if om.Len() != 0 {
			t.Errorf("Expected empty map length 0, got %d", om.Len())
		}

		if len(om.Keys()) != 0 {
			t.Errorf("Expected empty keys slice")
		}

		if len(om.Values()) != 0 {
			t.Errorf("Expected empty values slice")
		}

		if str := om.String(); str != "{}" {
			t.Errorf("Expected empty map string {}, got %s", str)
		}
	})
}

func TestOrderedMap_Range(t *testing.T) {
	om := NewOrderedMap()
	elements := []struct {
		key   string
		value int
	}{
		{"a", 1},
		{"b", 2},
		{"c", 3},
	}

	// Add elements
	for _, elem := range elements {
		om.Set(elem.key, elem.value)
	}

	// Test Range
	t.Run("Range All Elements", func(t *testing.T) {
		index := 0
		om.Range(func(key, value any) bool {
			if key != elements[index].key || value != elements[index].value {
				t.Errorf("Expected (%v, %v) at index %d, got (%v, %v)",
					elements[index].key, elements[index].value,
					index, key, value)
			}
			index++
			return true
		})
		if index != len(elements) {
			t.Errorf("Expected to iterate over %d elements, got %d", len(elements), index)
		}
	})

	// Test Range Early Stop
	t.Run("Range Early Stop", func(t *testing.T) {
		count := 0
		om.Range(func(key, value any) bool {
			count++
			return count < 2 // Stop after first element
		})
		if count != 2 {
			t.Errorf("Expected to stop after 2 elements, got %d", count)
		}
	})
}

func TestOrderedMap_Clear(t *testing.T) {
	om := NewOrderedMap()
	om.Set("key1", 1)
	om.Set("key2", 2)

	t.Run("Clear Map", func(t *testing.T) {
		om.Clear()
		if om.Len() != 0 {
			t.Errorf("Expected empty map after clear, got length %d", om.Len())
		}
		if len(om.Keys()) != 0 {
			t.Errorf("Expected no keys after clear")
		}
		if val, exists := om.Get("key1"); exists {
			t.Errorf("Expected no values after clear, got %v", val)
		}
	})
}

func TestOrderedMap_Copy(t *testing.T) {
	om := NewOrderedMap()
	om.Set("key1", 1)
	om.Set("key2", 2)

	t.Run("Copy Map", func(t *testing.T) {
		copy := om.Copy()

		// Check length
		if copy.Len() != om.Len() {
			t.Errorf("Expected copy to have same length")
		}

		// Check all elements
		om.Range(func(key, value any) bool {
			copyVal, exists := copy.Get(key)
			if !exists {
				t.Errorf("Key %v not found in copy", key)
				return false
			}
			if copyVal != value {
				t.Errorf("Value mismatch for key %v: expected %v, got %v", key, value, copyVal)
			}
			return true
		})

		// Verify independence
		om.Set("key3", 3)
		if _, exists := copy.Get("key3"); exists {
			t.Error("Copy should not be affected by changes to original")
		}
	})
}

func TestOrderedMap_Has(t *testing.T) {
	om := NewOrderedMap()
	om.Set("key1", 1)

	t.Run("Has Existing Key", func(t *testing.T) {
		if !om.Has("key1") {
			t.Error("Expected Has to return true for existing key")
		}
	})

	t.Run("Has Non-existing Key", func(t *testing.T) {
		if om.Has("nonexistent") {
			t.Error("Expected Has to return false for non-existing key")
		}
	})
}

func TestOrderedMap_ConcurrentRangeAndModify(t *testing.T) {
	om := NewOrderedMap()
	for i := 0; i < 100; i++ {
		om.Set(i, i)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrent Range
	go func() {
		defer wg.Done()
		om.Range(func(key, value any) bool {
			return true
		})
	}()

	// Concurrent Modification
	go func() {
		defer wg.Done()
		om.Set("new", 1000)
		om.Delete(50)
	}()

	wg.Wait()
}

func TestOrderedMap_JSONMarshaling(t *testing.T) {
	om := NewOrderedMap()

	// Test data structure
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Test data
	testData := map[string]interface{}{
		"user1": Person{Name: "John", Age: 30},
		"user2": Person{Name: "Mike", Age: 25},
		"settings": map[string]string{
			"theme": "dark",
			"lang":  "en",
		},
	}

	// Add data to OrderedMap
	for k, v := range testData {
		om.Set(k, v)
	}

	t.Run("Marshal to JSON", func(t *testing.T) {
		// Convert OrderedMap to standard map
		data := make(map[string]interface{})
		om.Range(func(key, value interface{}) bool {
			// Convert Person struct to JSON then to map
			if _, ok := value.(Person); ok {
				jsonBytes, err := json.Marshal(value)
				if err != nil {
					t.Errorf("Person marshal error: %v", err)
					return false
				}
				var personMap map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &personMap); err != nil {
					t.Errorf("Person unmarshal error: %v", err)
					return false
				}
				data[key.(string)] = personMap
			} else {
				data[key.(string)] = value
			}
			return true
		})

		// Convert to JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			t.Errorf("JSON marshal error: %v", err)
			return
		}

		// Convert JSON back to map
		var unmarshaledData map[string]interface{}
		if err := json.Unmarshal(jsonData, &unmarshaledData); err != nil {
			t.Errorf("JSON unmarshal error: %v", err)
			return
		}

		// Compare settings data
		settingsData, ok := unmarshaledData["settings"].(map[string]interface{})
		if !ok {
			t.Error("settings data is not of type map[string]interface{}")
			return
		}

		expectedSettings := testData["settings"].(map[string]string)
		if settingsData["theme"] != expectedSettings["theme"] ||
			settingsData["lang"] != expectedSettings["lang"] {
			t.Errorf("Settings data does not match.\nExpected: %v\nGot: %v",
				expectedSettings, settingsData)
		}

		// Compare User1 data
		user1Data, ok := unmarshaledData["user1"].(map[string]interface{})
		if !ok {
			t.Error("user1 data is not of type map[string]interface{}")
			return
		}

		expectedUser1 := testData["user1"].(Person)
		if user1Data["name"] != expectedUser1.Name ||
			int(user1Data["age"].(float64)) != expectedUser1.Age {
			t.Errorf("User1 data does not match.\nExpected: %v\nGot: %v",
				expectedUser1, user1Data)
		}
	})

	t.Run("Unmarshal from JSON", func(t *testing.T) {
		jsonStr := `{
			"user1": {"name": "Bob", "age": 35},
			"user2": {"name": "Alice", "age": 28},
			"settings": {"theme": "light", "lang": "en"}
		}`

		// Parse JSON
		var parsedData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsedData); err != nil {
			t.Errorf("JSON parse error: %v", err)
			return
		}

		// Create new OrderedMap and add parsed data
		newOm := NewOrderedMap()
		for k, v := range parsedData {
			newOm.Set(k, v)
		}

		// Check data
		if val, exists := newOm.Get("user1"); !exists {
			t.Error("user1 data not found")
		} else {
			userData := val.(map[string]interface{})
			if userData["name"] != "Bob" || userData["age"].(float64) != 35 {
				t.Errorf("user1 data is invalid: %v", userData)
			}
		}

		if val, exists := newOm.Get("settings"); !exists {
			t.Error("settings data not found")
		} else {
			settings := val.(map[string]interface{})
			if settings["theme"] != "light" || settings["lang"] != "en" {
				t.Errorf("settings data is invalid: %v", settings)
			}
		}
	})
}

func TestOrderedMap_EdgeCases(t *testing.T) {
	om := NewOrderedMap()

	t.Run("Nil Key Operations", func(t *testing.T) {
		// Set with nil key
		if err := om.Set(nil, "value"); err == nil {
			t.Error("Expected error when setting nil key")
		}

		// Get with nil key
		if _, exists := om.Get(nil); exists {
			t.Error("Expected false when getting nil key")
		}

		// Delete with nil key
		if err := om.Delete(nil); err == nil {
			t.Error("Expected error when deleting nil key")
		}

		// Has with nil key
		if om.Has(nil) {
			t.Error("Expected false when checking nil key")
		}
	})

	t.Run("Empty Map Operations", func(t *testing.T) {
		// Operations on empty map
		om.Clear()
		if om.head != nil || om.tail != nil {
			t.Error("Head and tail should be nil in empty map")
		}
		if len(om.nodeMap) != 0 {
			t.Error("NodeMap should be empty")
		}
	})
}

func BenchmarkOrderedMap(b *testing.B) {
	b.Run("Set", func(b *testing.B) {
		om := NewOrderedMap()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			om.Set(i, i)
		}
	})

	b.Run("Get", func(b *testing.B) {
		om := NewOrderedMap()
		for i := 0; i < 1000; i++ {
			om.Set(i, i)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			om.Get(i % 1000)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		om := NewOrderedMap()
		for i := 0; i < 1000; i++ {
			om.Set(i, i)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			om.Delete(i % 1000)
		}
	})

	b.Run("Range", func(b *testing.B) {
		om := NewOrderedMap()
		for i := 0; i < 1000; i++ {
			om.Set(i, i)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			om.Range(func(key, value any) bool {
				return true
			})
		}
	})
}

func TestOrderedMap_ConcurrentStress(t *testing.T) {
	om := NewOrderedMap()
	numOps := 1000
	numGoroutines := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 4) // 4 farklı operasyon tipi

	for i := 0; i < numGoroutines; i++ {
		// Setter
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				om.Set(base*numOps+j, j)
			}
		}(i)

		// Getter
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				om.Get(base*numOps + j)
			}
		}(i)

		// Deleter
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				om.Delete(base*numOps + j)
			}
		}(i)

		// Range
		go func() {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				om.Range(func(key, value any) bool {
					return true
				})
			}
		}()
	}

	wg.Wait()
}

func TestOrderedMap_DataConsistency(t *testing.T) {
	om := NewOrderedMap()

	t.Run("Order Consistency", func(t *testing.T) {
		// Add elements in specific order
		items := []string{"first", "second", "third"}
		for _, item := range items {
			om.Set(item, item)
		}

		// Delete middle element
		om.Delete("second")

		// Verify order
		keys := om.Keys()
		if len(keys) != 2 {
			t.Errorf("Expected 2 keys, got %d", len(keys))
		}
		if keys[0] != "first" || keys[1] != "third" {
			t.Error("Order not maintained after deletion")
		}

		// Add new element
		om.Set("fourth", "fourth")
		keys = om.Keys()
		if keys[2] != "fourth" {
			t.Error("New element not added at the end")
		}
	})
}

func TestComplexJSONMarshaling(t *testing.T) {
	// Define complex nested structures for JSON testing
	type Social struct {
		Platform string   `json:"platform"`
		URL      string   `json:"url"`
		Tags     []string `json:"tags"`
	}

	type Location struct {
		Street      string    `json:"street"`
		City        string    `json:"city"`
		Country     string    `json:"country"`
		Coordinates []float64 `json:"coordinates"`
	}

	type Contact struct {
		Type    string   `json:"type"`
		Value   string   `json:"value"`
		Primary bool     `json:"primary"`
		Hours   []string `json:"hours,omitempty"`
	}

	type Company struct {
		Name     string `json:"name"`
		Industry string `json:"industry"`
		Founded  int    `json:"founded"`
		Active   bool   `json:"active"`
	}

	type Person struct {
		ID          int                    `json:"id"`
		Name        string                 `json:"name"`
		Age         int                    `json:"age"`
		Email       string                 `json:"email"`
		Active      bool                   `json:"active"`
		Location    Location               `json:"location"`
		Contacts    []Contact              `json:"contacts"`
		SocialMedia []Social               `json:"social_media"`
		Company     Company                `json:"company"`
		Skills      []string               `json:"skills"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	// Create test data with complex nested structure
	person1 := Person{
		ID:     1,
		Name:   "Mesut Genez",
		Age:    30,
		Email:  "mesutgenez@gmail.com",
		Active: true,
		Location: Location{
			Street:      "Tech Valley 123",
			City:        "Istanbul",
			Country:     "Turkey",
			Coordinates: []float64{41.0082, 28.9784},
		},
		Contacts: []Contact{
			{
				Type:    "phone",
				Value:   "+90 555 123 4567",
				Primary: true,
				Hours:   []string{"09:00-18:00", "Mon-Fri"},
			},
			{
				Type:    "email",
				Value:   "work@example.com",
				Primary: false,
			},
		},
		SocialMedia: []Social{
			{
				Platform: "GitHub",
				URL:      "https://github.com/mstgnz",
				Tags:     []string{"go", "developer", "open-source"},
			},
			{
				Platform: "LinkedIn",
				URL:      "https://linkedin.com/in/mesutgenez",
				Tags:     []string{"software", "engineering"},
			},
		},
		Company: Company{
			Name:     "Tech Corp",
			Industry: "Software",
			Founded:  2020,
			Active:   true,
		},
		Skills: []string{"Go", "Python", "Docker", "Kubernetes"},
		Metadata: map[string]interface{}{
			"last_login":  "2024-01-02T15:04:05Z",
			"login_count": 42,
			"preferences": map[string]interface{}{
				"theme":         "dark",
				"timezone":      "Europe/Istanbul",
				"notifications": true,
			},
		},
	}

	// Test 1: Marshal complex structure
	om := NewOrderedMap()
	err := om.Set("person1", person1)
	if err != nil {
		t.Errorf("Failed to set person1: %v", err)
	}

	jsonData, err := json.MarshalIndent(om, "", "    ")
	if err != nil {
		t.Errorf("Failed to marshal: %v", err)
	}

	// Test 2: Unmarshal complex structure
	newMap := NewOrderedMap()
	err = json.Unmarshal(jsonData, newMap)
	if err != nil {
		t.Errorf("Failed to unmarshal: %v", err)
	}

	// Verify data integrity
	value, exists := newMap.Get("person1")
	if !exists {
		t.Error("Failed to get person1 from unmarshaled data")
	}

	// Convert to map and verify values
	personMap, ok := value.(map[string]interface{})
	if !ok {
		t.Error("Failed to convert person1 to map")
	}

	// Check basic fields
	if personMap["name"] != person1.Name {
		t.Errorf("Name mismatch. Expected %v, got %v", person1.Name, personMap["name"])
	}

	if personMap["email"] != person1.Email {
		t.Errorf("Email mismatch. Expected %v, got %v", person1.Email, personMap["email"])
	}

	// Check nested structures
	location, ok := personMap["location"].(map[string]interface{})
	if !ok {
		t.Error("Failed to get location data")
	} else {
		if location["city"] != person1.Location.City {
			t.Errorf("City mismatch. Expected %v, got %v", person1.Location.City, location["city"])
		}
	}

	// Check array structures
	skills, ok := personMap["skills"].([]interface{})
	if !ok {
		t.Error("Failed to get skills array")
	} else {
		if len(skills) != len(person1.Skills) {
			t.Errorf("Skills length mismatch. Expected %v, got %v", len(person1.Skills), len(skills))
		}
	}

	// Check deeply nested structures like metadata
	metadata, ok := personMap["metadata"].(map[string]interface{})
	if !ok {
		t.Error("Failed to get metadata")
	} else {
		preferences, ok := metadata["preferences"].(map[string]interface{})
		if !ok {
			t.Error("Failed to get preferences from metadata")
		} else {
			if preferences["theme"] != "dark" {
				t.Errorf("Theme preference mismatch. Expected 'dark', got %v", preferences["theme"])
			}
		}
	}
}

func TestOrderedMap_MarshalJSONEdgeCases(t *testing.T) {
	om := NewOrderedMap()

	// Test with non-string key
	om.Set(123, "value")
	om.Set(true, "bool-value")
	om.Set(3.14, "float-value")

	jsonData, err := json.Marshal(om)
	if err != nil {
		t.Errorf("Failed to marshal with non-string keys: %v", err)
	}

	// Verify the marshaled data
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Errorf("Failed to unmarshal test data: %v", err)
	}

	// Check if non-string keys were converted to strings
	if result["123"] != "value" {
		t.Error("Integer key not properly marshaled")
	}
	if result["true"] != "bool-value" {
		t.Error("Boolean key not properly marshaled")
	}
	if result["3.14"] != "float-value" {
		t.Error("Float key not properly marshaled")
	}
}

func TestOrderedMap_UnmarshalJSONEdgeCases(t *testing.T) {
	// Test with invalid JSON
	om := NewOrderedMap()
	err := json.Unmarshal([]byte(`{invalid json}`), om)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test with empty JSON object
	err = json.Unmarshal([]byte(`{}`), om)
	if err != nil {
		t.Errorf("Failed to unmarshal empty object: %v", err)
	}
	if om.Len() != 0 {
		t.Error("Expected empty map after unmarshaling empty object")
	}

	// Test with null values
	err = json.Unmarshal([]byte(`{"key": null}`), om)
	if err != nil {
		t.Errorf("Failed to unmarshal null value: %v", err)
	}
	val, exists := om.Get("key")
	if !exists || val != nil {
		t.Error("Expected null value to be stored as nil")
	}

	// Test with complex nested structure
	jsonStr := `{
		"array": [1, 2, 3],
		"nested": {
			"a": 1,
			"b": "string",
			"c": true,
			"d": null,
			"e": [{"x": 1}]
		}
	}`
	err = json.Unmarshal([]byte(jsonStr), om)
	if err != nil {
		t.Errorf("Failed to unmarshal complex structure: %v", err)
	}

	// Verify array
	array, exists := om.Get("array")
	if !exists {
		t.Error("Failed to get array")
	}
	arrayVal, ok := array.([]interface{})
	if !ok || len(arrayVal) != 3 {
		t.Error("Array not properly unmarshaled")
	}

	// Verify nested structure
	nested, exists := om.Get("nested")
	if !exists {
		t.Error("Failed to get nested structure")
	}
	nestedMap, ok := nested.(map[string]interface{})
	if !ok {
		t.Error("Nested structure not properly unmarshaled")
	}
	if nestedMap["a"] != float64(1) || nestedMap["b"] != "string" || nestedMap["c"] != true || nestedMap["d"] != nil {
		t.Error("Nested values not properly unmarshaled")
	}
}

func TestOrderedMap_SetEdgeCases(t *testing.T) {
	om := NewOrderedMap()

	// Test setting nil value
	err := om.Set("key", nil)
	if err != nil {
		t.Errorf("Failed to set nil value: %v", err)
	}
	val, exists := om.Get("key")
	if !exists || val != nil {
		t.Error("Nil value not properly stored")
	}

	// Test setting complex nested structures
	type nested struct {
		Field string
	}
	complexValue := map[string]interface{}{
		"array":   []int{1, 2, 3},
		"string":  "value",
		"number":  42,
		"boolean": true,
		"null":    nil,
		"struct":  nested{Field: "test"},
		"map": map[string]int{
			"one": 1,
			"two": 2,
		},
	}

	err = om.Set("complex", complexValue)
	if err != nil {
		t.Errorf("Failed to set complex value: %v", err)
	}

	// Verify the complex value was stored correctly
	retrieved, exists := om.Get("complex")
	if !exists {
		t.Error("Failed to get complex value")
	}

	retrievedMap, ok := retrieved.(map[string]interface{})
	if !ok {
		t.Error("Complex value not stored as map")
	}

	// Verify map contents
	if len(retrievedMap) != len(complexValue) {
		t.Error("Complex value not stored with all fields")
	}

	// Test updating existing key with different type
	om.Set("key", "string")
	om.Set("key", 123)
	om.Set("key", true)
	val, _ = om.Get("key")
	if val != true {
		t.Error("Failed to update existing key with different type")
	}
}

func TestOrderedMap_UnmarshalJSONComplete(t *testing.T) {
	om := NewOrderedMap()

	// Test with various JSON types and structures
	jsonStr := `{
		"nil_value": null,
		"number_int": 42,
		"number_float": 3.14,
		"string": "hello",
		"boolean": true,
		"array": [1, "two", true, null, {"nested": "object"}],
		"object": {
			"a": 1,
			"b": "2",
			"c": true,
			"d": null,
			"e": [1, 2, 3],
			"f": {"nested": "value"}
		},
		"empty_object": {},
		"empty_array": []
	}`

	// Test unmarshaling
	err := json.Unmarshal([]byte(jsonStr), om)
	if err != nil {
		t.Errorf("Failed to unmarshal complete test data: %v", err)
	}

	// Verify all values were stored correctly
	testCases := []struct {
		key      string
		checkFn  func(interface{}) bool
		errorMsg string
	}{
		{
			key: "nil_value",
			checkFn: func(v interface{}) bool {
				return v == nil
			},
			errorMsg: "nil value not properly stored",
		},
		{
			key: "number_int",
			checkFn: func(v interface{}) bool {
				num, ok := v.(float64)
				return ok && num == 42
			},
			errorMsg: "integer value not properly stored",
		},
		{
			key: "number_float",
			checkFn: func(v interface{}) bool {
				num, ok := v.(float64)
				return ok && num == 3.14
			},
			errorMsg: "float value not properly stored",
		},
		{
			key: "string",
			checkFn: func(v interface{}) bool {
				str, ok := v.(string)
				return ok && str == "hello"
			},
			errorMsg: "string value not properly stored",
		},
		{
			key: "boolean",
			checkFn: func(v interface{}) bool {
				b, ok := v.(bool)
				return ok && b == true
			},
			errorMsg: "boolean value not properly stored",
		},
		{
			key: "array",
			checkFn: func(v interface{}) bool {
				arr, ok := v.([]interface{})
				return ok && len(arr) == 5
			},
			errorMsg: "array not properly stored",
		},
		{
			key: "object",
			checkFn: func(v interface{}) bool {
				obj, ok := v.(map[string]interface{})
				return ok && len(obj) == 6
			},
			errorMsg: "object not properly stored",
		},
		{
			key: "empty_object",
			checkFn: func(v interface{}) bool {
				obj, ok := v.(map[string]interface{})
				return ok && len(obj) == 0
			},
			errorMsg: "empty object not properly stored",
		},
		{
			key: "empty_array",
			checkFn: func(v interface{}) bool {
				arr, ok := v.([]interface{})
				return ok && len(arr) == 0
			},
			errorMsg: "empty array not properly stored",
		},
	}

	for _, tc := range testCases {
		value, exists := om.Get(tc.key)
		if !exists {
			t.Errorf("Key not found: %s", tc.key)
			continue
		}
		if !tc.checkFn(value) {
			t.Error(tc.errorMsg)
		}
	}

	// Test marshaling back
	jsonData, err := json.Marshal(om)
	if err != nil {
		t.Errorf("Failed to marshal back to JSON: %v", err)
	}

	// Unmarshal to verify structure
	var verifyMap map[string]interface{}
	err = json.Unmarshal(jsonData, &verifyMap)
	if err != nil {
		t.Errorf("Failed to unmarshal verification data: %v", err)
	}

	// Verify the number of keys
	if len(verifyMap) != 9 {
		t.Errorf("Expected 9 keys in marshaled data, got %d", len(verifyMap))
	}
}

func TestOrderedMap_SetComplete(t *testing.T) {
	om := NewOrderedMap()

	// Test nil key
	err := om.set(nil, "value")
	if err == nil {
		t.Error("Expected error for nil key")
	}

	// Test setting first element
	err = om.set("first", "value1")
	if err != nil {
		t.Errorf("Failed to set first element: %v", err)
	}
	if om.head.Key != "first" || om.tail.Key != "first" {
		t.Error("Head and tail not properly set for first element")
	}

	// Test setting second element
	err = om.set("second", "value2")
	if err != nil {
		t.Errorf("Failed to set second element: %v", err)
	}
	if om.head.Key != "first" || om.tail.Key != "second" {
		t.Error("Head and tail not properly set after second element")
	}

	// Test updating existing value
	err = om.set("first", "updated")
	if err != nil {
		t.Errorf("Failed to update existing value: %v", err)
	}
	if val, _ := om.Get("first"); val != "updated" {
		t.Error("Value not properly updated")
	}

	// Test that the order is maintained after update
	keys := om.Keys()
	if len(keys) != 2 || keys[0] != "first" || keys[1] != "second" {
		t.Error("Order not maintained after update")
	}

	// Test setting various types of values
	testCases := []struct {
		key   string
		value interface{}
	}{
		{"nil", nil},
		{"int", 42},
		{"float", 3.14},
		{"string", "test"},
		{"bool", true},
	}

	for _, tc := range testCases {
		err := om.set(tc.key, tc.value)
		if err != nil {
			t.Errorf("Failed to set %v: %v", tc.key, err)
		}
		val, exists := om.Get(tc.key)
		if !exists {
			t.Errorf("Key %v not found after set", tc.key)
		}
		if val != tc.value {
			t.Errorf("Value mismatch for key %v", tc.key)
		}
	}

	// Test setting complex types (without direct comparison)
	complexCases := []struct {
		key   string
		value interface{}
	}{
		{"slice", []int{1, 2, 3}},
		{"map", map[string]string{"key": "value"}},
	}

	for _, tc := range complexCases {
		err := om.set(tc.key, tc.value)
		if err != nil {
			t.Errorf("Failed to set %v: %v", tc.key, err)
		}
		val, exists := om.Get(tc.key)
		if !exists {
			t.Errorf("Key %v not found after set", tc.key)
		}
		// Type check only for complex types
		switch v := tc.value.(type) {
		case []int:
			if slice, ok := val.([]int); !ok || len(slice) != len(v) {
				t.Errorf("Slice value type or length mismatch for key %v", tc.key)
			}
		case map[string]string:
			if m, ok := val.(map[string]string); !ok || len(m) != len(v) {
				t.Errorf("Map value type or length mismatch for key %v", tc.key)
			}
		}
	}
}

func TestOrderedMap_UnmarshalJSONError(t *testing.T) {
	om := NewOrderedMap()

	// Test with nil data
	err := om.UnmarshalJSON(nil)
	if err == nil {
		t.Error("Expected error for nil data")
	}

	// Test with invalid JSON
	invalidJSON := []byte(`{"key": invalid}`)
	err = om.UnmarshalJSON(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test with non-object JSON
	nonObjectJSON := []byte(`["array"]`)
	err = om.UnmarshalJSON(nonObjectJSON)
	if err == nil {
		t.Error("Expected error for non-object JSON")
	}

	// Test with empty JSON object
	emptyJSON := []byte(`{}`)
	err = om.UnmarshalJSON(emptyJSON)
	if err != nil {
		t.Errorf("Unexpected error for empty object: %v", err)
	}
	if om.Len() != 0 {
		t.Error("Map should be empty after unmarshaling empty object")
	}

	// Test with single null value
	nullJSON := []byte(`{"key": null}`)
	err = om.UnmarshalJSON(nullJSON)
	if err != nil {
		t.Errorf("Unexpected error for null value: %v", err)
	}
	if val, exists := om.Get("key"); !exists || val != nil {
		t.Error("Null value not properly unmarshaled")
	}
}

func TestOrderedMap_FirstLast(t *testing.T) {
	om := NewOrderedMap()

	// Test empty map
	t.Run("Empty Map", func(t *testing.T) {
		key, value, exists := om.First()
		if exists || key != nil || value != nil {
			t.Error("Expected First() to return nil, nil, false for empty map")
		}

		key, value, exists = om.Last()
		if exists || key != nil || value != nil {
			t.Error("Expected Last() to return nil, nil, false for empty map")
		}
	})

	// Add elements
	elements := []struct {
		key   string
		value int
	}{
		{"first", 1},
		{"second", 2},
		{"third", 3},
	}

	for _, elem := range elements {
		om.Set(elem.key, elem.value)
	}

	// Test First
	t.Run("First Element", func(t *testing.T) {
		key, value, exists := om.First()
		if !exists {
			t.Error("Expected First() to return true for non-empty map")
		}
		if key != "first" || value != 1 {
			t.Errorf("Expected first element to be (first, 1), got (%v, %v)", key, value)
		}
	})

	// Test Last
	t.Run("Last Element", func(t *testing.T) {
		key, value, exists := om.Last()
		if !exists {
			t.Error("Expected Last() to return true for non-empty map")
		}
		if key != "third" || value != 3 {
			t.Errorf("Expected last element to be (third, 3), got (%v, %v)", key, value)
		}
	})
}

func TestOrderedMap_Reverse(t *testing.T) {
	om := NewOrderedMap()

	// Test empty map
	t.Run("Empty Map", func(t *testing.T) {
		reversed := om.Reverse()
		if reversed.Len() != 0 {
			t.Error("Expected reversed empty map to be empty")
		}
	})

	// Add elements
	elements := []struct {
		key   string
		value int
	}{
		{"a", 1},
		{"b", 2},
		{"c", 3},
	}

	for _, elem := range elements {
		om.Set(elem.key, elem.value)
	}

	// Test reverse
	t.Run("Reverse Order", func(t *testing.T) {
		reversed := om.Reverse()
		if reversed.Len() != len(elements) {
			t.Errorf("Expected reversed map to have %d elements, got %d", len(elements), reversed.Len())
		}

		keys := reversed.Keys()
		for i := range elements {
			expectedKey := elements[len(elements)-1-i].key
			if keys[i] != expectedKey {
				t.Errorf("Expected key %s at position %d, got %v", expectedKey, i, keys[i])
			}
		}
	})
}

func TestOrderedMap_Filter(t *testing.T) {
	om := NewOrderedMap()

	// Add elements
	for i := 1; i <= 5; i++ {
		om.Set(fmt.Sprintf("key%d", i), i)
	}

	// Test filtering even numbers
	t.Run("Filter Even Numbers", func(t *testing.T) {
		filtered := om.Filter(func(key, value any) bool {
			if val, ok := value.(int); ok {
				return val%2 == 0
			}
			return false
		})

		if filtered.Len() != 2 {
			t.Errorf("Expected 2 even numbers, got %d", filtered.Len())
		}

		values := filtered.Values()
		for _, v := range values {
			if val, ok := v.(int); !ok || val%2 != 0 {
				t.Errorf("Expected even number, got %v", v)
			}
		}
	})

	// Test filtering with empty result
	t.Run("Filter No Match", func(t *testing.T) {
		filtered := om.Filter(func(key, value any) bool {
			return false
		})

		if filtered.Len() != 0 {
			t.Error("Expected empty filtered map")
		}
	})
}

func TestOrderedMap_Map(t *testing.T) {
	om := NewOrderedMap()

	// Add elements
	for i := 1; i <= 3; i++ {
		om.Set(fmt.Sprintf("key%d", i), i)
	}

	// Test doubling values
	t.Run("Double Values", func(t *testing.T) {
		doubled := om.Map(func(key, value any) (any, any) {
			if val, ok := value.(int); ok {
				return key, val * 2
			}
			return key, value
		})

		if doubled.Len() != om.Len() {
			t.Errorf("Expected %d elements, got %d", om.Len(), doubled.Len())
		}

		values := doubled.Values()
		for i, v := range values {
			expected := (i + 1) * 2
			if val, ok := v.(int); !ok || val != expected {
				t.Errorf("Expected %d, got %v", expected, v)
			}
		}
	})

	// Test key transformation
	t.Run("Transform Keys", func(t *testing.T) {
		transformed := om.Map(func(key, value any) (any, any) {
			return fmt.Sprintf("prefix_%v", key), value
		})

		keys := transformed.Keys()
		for _, k := range keys {
			if key, ok := k.(string); !ok || len(key) <= 6 || key[:7] != "prefix_" {
				t.Errorf("Expected key with 'prefix_' prefix, got %v", k)
			}
		}
	})
}

func TestOrderedMap_JSONOperations(t *testing.T) {
	om := NewOrderedMap()

	// Add test data
	testData := map[string]interface{}{
		"string": "value",
		"int":    42,
		"float":  3.14,
		"bool":   true,
	}

	for k, v := range testData {
		om.Set(k, v)
	}

	// Test ToJSON with different options
	t.Run("ToJSON Basic", func(t *testing.T) {
		data, err := om.ToJSON(nil)
		if err != nil {
			t.Errorf("ToJSON failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("Expected non-empty JSON data")
		}
	})

	t.Run("ToJSON PrettyPrint", func(t *testing.T) {
		opts := &JSONOptions{
			KeyAsString: true,
			PrettyPrint: true,
		}
		data, err := om.ToJSON(opts)
		if err != nil {
			t.Errorf("ToJSON with PrettyPrint failed: %v", err)
		}
		if !bytes.Contains(data, []byte("\n")) {
			t.Error("Expected pretty printed JSON to contain newlines")
		}
	})

	// Test FromJSON with different options
	t.Run("FromJSON Basic", func(t *testing.T) {
		jsonData := []byte(`{"key1": 123, "key2": "value2"}`)
		newMap := NewOrderedMap()
		err := newMap.FromJSON(jsonData, nil)
		if err != nil {
			t.Errorf("FromJSON failed: %v", err)
		}
		if newMap.Len() != 2 {
			t.Errorf("Expected 2 elements, got %d", newMap.Len())
		}
	})

	t.Run("FromJSON PreserveType", func(t *testing.T) {
		jsonData := []byte(`{"int": 42, "float": 3.14}`)
		newMap := NewOrderedMap()
		opts := &JSONOptions{
			PreserveType: true,
		}
		err := newMap.FromJSON(jsonData, opts)
		if err != nil {
			t.Errorf("FromJSON with PreserveType failed: %v", err)
		}

		// Check if types are preserved
		if val, ok := newMap.Get("int"); !ok {
			t.Error("Expected to find 'int' key")
		} else if _, ok := val.(float64); !ok {
			t.Errorf("Expected float64 type for 'int' value, got %T", val)
		}

		if val, ok := newMap.Get("float"); !ok {
			t.Error("Expected to find 'float' key")
		} else if _, ok := val.(float64); !ok {
			t.Errorf("Expected float64 type for 'float' value, got %T", val)
		}
	})

	t.Run("FromJSON Invalid", func(t *testing.T) {
		invalidJSON := []byte(`{"key": invalid}`)
		newMap := NewOrderedMap()
		err := newMap.FromJSON(invalidJSON, nil)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestOrderedMap_ToJSONComplete(t *testing.T) {
	om := NewOrderedMap()

	// Test with nil options
	t.Run("Nil Options", func(t *testing.T) {
		data, err := om.ToJSON(nil)
		if err != nil {
			t.Errorf("ToJSON with nil options failed: %v", err)
		}
		if !json.Valid(data) {
			t.Error("Invalid JSON output")
		}
	})

	// Test with non-string keys
	t.Run("Non-String Keys", func(t *testing.T) {
		om.Set(123, "value")
		om.Set(true, "bool-value")

		opts := &JSONOptions{
			KeyAsString: false,
		}
		_, err := om.ToJSON(opts)
		if err == nil {
			t.Error("Expected error for non-string keys with KeyAsString=false")
		}
	})

	// Test with preserve type
	t.Run("Preserve Type", func(t *testing.T) {
		om.Clear()
		om.Set("int", "42")
		om.Set("float", "3.14")

		opts := &JSONOptions{
			PreserveType: true,
		}
		data, err := om.ToJSON(opts)
		if err != nil {
			t.Errorf("ToJSON with preserve type failed: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Errorf("Failed to unmarshal result: %v", err)
		}
	})

	// Test pretty print with complex data
	t.Run("Pretty Print Complex", func(t *testing.T) {
		om.Clear()
		complexData := map[string]interface{}{
			"array": []int{1, 2, 3},
			"nested": map[string]string{
				"key": "value",
			},
		}
		om.Set("complex", complexData)

		opts := &JSONOptions{
			PrettyPrint: true,
		}
		data, err := om.ToJSON(opts)
		if err != nil {
			t.Errorf("ToJSON with pretty print failed: %v", err)
		}
		if !bytes.Contains(data, []byte("\n")) {
			t.Error("Pretty printed JSON should contain newlines")
		}
	})
}

func TestOrderedMap_FromJSONComplete(t *testing.T) {
	// Test with various numeric types
	t.Run("Numeric Types", func(t *testing.T) {
		jsonData := []byte(`{
			"int": 42,
			"float": 3.14,
			"scientific": 1.23e5
		}`)

		om := NewOrderedMap()
		opts := &JSONOptions{
			PreserveType: true,
		}
		err := om.FromJSON(jsonData, opts)
		if err != nil {
			t.Errorf("FromJSON failed: %v", err)
		}

		// Check int conversion
		if val, exists := om.Get("int"); !exists {
			t.Error("int key not found")
		} else if _, ok := val.(float64); !ok {
			t.Errorf("Expected float64 type for 'int', got %T", val)
		}

		// Check float conversion
		if val, exists := om.Get("float"); !exists {
			t.Error("float key not found")
		} else if f, ok := val.(float64); !ok || f != 3.14 {
			t.Errorf("Expected float64 value 3.14, got %v of type %T", val, val)
		}

		// Check scientific notation
		if val, exists := om.Get("scientific"); !exists {
			t.Error("scientific key not found")
		} else if f, ok := val.(float64); !ok || f != 1.23e5 {
			t.Errorf("Expected float64 value 1.23e5, got %v of type %T", val, val)
		}
	})

	// Test with non-string keys
	t.Run("Non-String Keys", func(t *testing.T) {
		jsonData := []byte(`{
			"123": "numeric key",
			"3.14": "float key",
			"true": "bool key"
		}`)

		om := NewOrderedMap()
		opts := &JSONOptions{
			KeyAsString: false,
		}
		err := om.FromJSON(jsonData, opts)
		if err != nil {
			t.Errorf("FromJSON failed: %v", err)
		}

		// Verify key conversions
		if _, exists := om.Get(int64(123)); !exists {
			t.Error("Numeric key not converted properly")
		}
		if _, exists := om.Get(float64(3.14)); !exists {
			t.Error("Float key not converted properly")
		}
	})

	// Test with invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		invalidCases := []string{
			`{"key": value}`,  // Missing quotes
			`{"key": 123,}`,   // Trailing comma
			`{"key"`,          // Incomplete
			`not json at all`, // Not JSON
		}

		for _, invalid := range invalidCases {
			om := NewOrderedMap()
			err := om.FromJSON([]byte(invalid), nil)
			if err == nil {
				t.Errorf("Expected error for invalid JSON: %s", invalid)
			}
		}
	})
}

func TestOrderedMap_UnmarshalJSONExtended(t *testing.T) {
	// Test with various JSON types
	t.Run("Various Types", func(t *testing.T) {
		jsonData := []byte(`{
			"null": null,
			"bool": true,
			"int": 42,
			"float": 3.14,
			"string": "hello",
			"array": [1,2,3],
			"object": {"key": "value"}
		}`)

		om := NewOrderedMap()
		err := om.UnmarshalJSON(jsonData)
		if err != nil {
			t.Errorf("UnmarshalJSON failed: %v", err)
		}

		// Check each type
		testCases := []struct {
			key  string
			test func(interface{}) bool
		}{
			{"null", func(v interface{}) bool { return v == nil }},
			{"bool", func(v interface{}) bool { b, ok := v.(bool); return ok && b }},
			{"int", func(v interface{}) bool { _, ok := v.(float64); return ok }},
			{"float", func(v interface{}) bool { _, ok := v.(float64); return ok }},
			{"string", func(v interface{}) bool { _, ok := v.(string); return ok }},
			{"array", func(v interface{}) bool { _, ok := v.([]interface{}); return ok }},
			{"object", func(v interface{}) bool { _, ok := v.(map[string]interface{}); return ok }},
		}

		for _, tc := range testCases {
			if val, exists := om.Get(tc.key); !exists || !tc.test(val) {
				t.Errorf("Invalid value for key %s: %v", tc.key, val)
			}
		}
	})

	// Test with empty and whitespace JSON
	t.Run("Empty and Whitespace", func(t *testing.T) {
		cases := []string{
			"{}",
			"  {  }  ",
			"{\n\t\r}",
		}

		for _, c := range cases {
			om := NewOrderedMap()
			err := om.UnmarshalJSON([]byte(c))
			if err != nil {
				t.Errorf("UnmarshalJSON failed for %q: %v", c, err)
			}
			if om.Len() != 0 {
				t.Errorf("Expected empty map for %q", c)
			}
		}
	})

	// Test with nested structures
	t.Run("Nested Structures", func(t *testing.T) {
		jsonData := []byte(`{
			"nested": {
				"array": [
					{"key": "value"},
					{"number": 42}
				],
				"map": {
					"deep": {
						"deeper": true
					}
				}
			}
		}`)

		om := NewOrderedMap()
		err := om.UnmarshalJSON(jsonData)
		if err != nil {
			t.Errorf("UnmarshalJSON failed: %v", err)
		}

		val, exists := om.Get("nested")
		if !exists {
			t.Fatal("Nested key not found")
		}

		nested, ok := val.(map[string]interface{})
		if !ok {
			t.Fatal("Nested value is not a map")
		}

		// Check array
		arr, ok := nested["array"].([]interface{})
		if !ok || len(arr) != 2 {
			t.Error("Array not properly unmarshaled")
		}

		// Check deep nesting
		if m, ok := nested["map"].(map[string]interface{}); ok {
			if deep, ok := m["deep"].(map[string]interface{}); ok {
				if _, ok := deep["deeper"].(bool); !ok {
					t.Error("Deep nesting not properly unmarshaled")
				}
			} else {
				t.Error("Deep nesting structure invalid")
			}
		} else {
			t.Error("Map structure invalid")
		}
	})
}
