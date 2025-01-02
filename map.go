package orderedmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

// Node represents a node in the doubly linked list that maintains the order of elements.
// Each node contains a key-value pair and pointers to the previous and next nodes.
type Node struct {
	Key   any   // The key of the key-value pair
	Value any   // The value associated with the key
	prev  *Node // Pointer to the previous node
	next  *Node // Pointer to the next node
}

// OrderedMap is a thread-safe implementation of an ordered map data structure.
// It combines a doubly linked list for maintaining insertion order with a hash map
// for O(1) lookups. All operations are protected by a read-write mutex for thread safety.
type OrderedMap struct {
	mu      sync.RWMutex  // Protects concurrent access to the map
	head    *Node         // Points to the first node in the list
	tail    *Node         // Points to the last node in the list
	nodeMap map[any]*Node // Maps keys to their corresponding nodes
	length  int           // Number of elements in the map
}

// NewOrderedMap creates and initializes a new empty OrderedMap.
// The returned map is ready to use and is thread-safe.
//
// Example:
//
//	om := NewOrderedMap()
//	om.Set("key", "value")
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		nodeMap: make(map[any]*Node),
	}
}

// Set adds a new key-value pair to the map or updates an existing one.
// If the key already exists, its value is updated. If the key is new,
// the pair is added to the end of the ordered list.
//
// The method is thread-safe and returns an error if the key is nil.
//
// Example:
//
//	om := NewOrderedMap()
//	err := om.Set("key", "value")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (om *OrderedMap) Set(key, value any) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	if node, exists := om.nodeMap[key]; exists {
		node.Value = value
		return nil
	}

	newNode := &Node{
		Key:   key,
		Value: value,
	}

	if om.tail == nil {
		om.head = newNode
		om.tail = newNode
	} else {
		newNode.prev = om.tail
		om.tail.next = newNode
		om.tail = newNode
	}

	om.nodeMap[key] = newNode
	om.length++
	return nil
}

// Delete removes the element with the given key from the map.
// If the key doesn't exist, the operation is a no-op and returns nil.
// The method is thread-safe and returns an error if the key is nil.
//
// Example:
//
//	err := om.Delete("key")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (om *OrderedMap) Delete(key any) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	node, exists := om.nodeMap[key]
	if !exists {
		return nil
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		om.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		om.tail = node.prev
	}

	delete(om.nodeMap, key)
	om.length--

	// Help GC by removing references
	node.prev = nil
	node.next = nil
	return nil
}

// Keys returns a slice containing all keys in the map in their insertion order.
// The returned slice is a copy of the keys, so modifications to the slice
// won't affect the map.
//
// Example:
//
//	keys := om.Keys()
//	for _, key := range keys {
//	    fmt.Println(key)
//	}
func (om *OrderedMap) Keys() []any {
	om.mu.RLock()
	defer om.mu.RUnlock()

	keys := make([]any, 0, om.length)
	current := om.head
	for current != nil {
		keys = append(keys, current.Key)
		current = current.next
	}
	return keys
}

// Values returns a slice containing all values in the map in their insertion order.
// The returned slice is a copy of the values, so modifications to the slice
// won't affect the map.
//
// Example:
//
//	values := om.Values()
//	for _, value := range values {
//	    fmt.Println(value)
//	}
func (om *OrderedMap) Values() []any {
	om.mu.RLock()
	defer om.mu.RUnlock()

	values := make([]any, 0, om.length)
	current := om.head
	for current != nil {
		values = append(values, current.Value)
		current = current.next
	}
	return values
}

// Range iterates over the map in insertion order and calls the given function
// for each key-value pair. If the function returns false, iteration stops.
// The method is thread-safe and holds a read lock during iteration.
//
// Example:
//
//	om.Range(func(key, value any) bool {
//	    fmt.Printf("%v: %v\n", key, value)
//	    return true // continue iteration
//	})
func (om *OrderedMap) Range(f func(key, value any) bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	current := om.head
	for current != nil {
		if !f(current.Key, current.Value) {
			break
		}
		current = current.next
	}
}

// Clear removes all elements from the map, resetting it to an empty state.
// This operation is not atomic - if you need atomicity, you should implement
// your own locking around this method.
//
// Example:
//
//	om.Clear()
func (om *OrderedMap) Clear() {
	om.nodeMap = make(map[any]*Node)
	om.head = nil
	om.tail = nil
	om.length = 0
}

// Get retrieves the value associated with the given key.
// Returns the value and true if the key exists, nil and false otherwise.
// The method is thread-safe and returns nil, false if the key is nil.
//
// Example:
//
//	if value, exists := om.Get("key"); exists {
//	    fmt.Printf("Value: %v\n", value)
//	}
func (om *OrderedMap) Get(key any) (any, bool) {
	if key == nil {
		return nil, false
	}

	om.mu.RLock()
	defer om.mu.RUnlock()

	if node, exists := om.nodeMap[key]; exists {
		return node.Value, true
	}
	return nil, false
}

// String returns a string representation of the map in the format {key1: value1, key2: value2}.
// The elements are ordered according to their insertion order.
// This method is thread-safe.
//
// Example:
//
//	fmt.Println(om.String()) // Output: {key1: value1, key2: value2}
func (om *OrderedMap) String() string {
	om.mu.RLock()
	defer om.mu.RUnlock()

	result := "{"
	current := om.head
	for current != nil {
		if current != om.head {
			result += ", "
		}
		result += fmt.Sprintf("%v: %v", current.Key, current.Value)
		current = current.next
	}
	result += "}"
	return result
}

// Len returns the number of elements in the map.
// This method is thread-safe.
//
// Example:
//
//	count := om.Len()
//	fmt.Printf("Map contains %d elements\n", count)
func (om *OrderedMap) Len() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.length
}

// Has checks if a key exists in the map.
// Returns true if the key exists, false otherwise.
// The method is thread-safe and returns false if the key is nil.
//
// Example:
//
//	if om.Has("key") {
//	    fmt.Println("Key exists")
//	}
func (om *OrderedMap) Has(key any) bool {
	if key == nil {
		return false
	}

	om.mu.RLock()
	defer om.mu.RUnlock()

	_, exists := om.nodeMap[key]
	return exists
}

// Copy creates a deep copy of the OrderedMap.
// The new map contains copies of all key-value pairs in the same order.
// This method is thread-safe.
//
// Example:
//
//	newMap := om.Copy()
func (om *OrderedMap) Copy() *OrderedMap {
	om.mu.RLock()
	defer om.mu.RUnlock()

	newMap := NewOrderedMap()
	current := om.head
	for current != nil {
		_ = newMap.Set(current.Key, current.Value)
		current = current.next
	}
	return newMap
}

// MarshalJSON implements the json.Marshaler interface.
// It converts the OrderedMap to a JSON object, maintaining the order of keys.
// Keys are converted to strings in the JSON representation.
//
// Example:
//
//	data, err := json.Marshal(om)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (om *OrderedMap) MarshalJSON() ([]byte, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	// Create a temporary map for JSON marshaling
	tmpMap := make(map[string]interface{})

	// Iterate through the ordered map and add to temporary map
	current := om.head
	for current != nil {
		// Convert key to string if possible
		keyStr, ok := current.Key.(string)
		if !ok {
			keyStr = fmt.Sprintf("%v", current.Key)
		}
		tmpMap[keyStr] = current.Value
		current = current.next
	}

	return json.Marshal(tmpMap)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It populates the OrderedMap from a JSON object, maintaining the order of keys
// as they appear in the JSON input.
//
// Example:
//
//	var om OrderedMap
//	err := json.Unmarshal(data, &om)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (om *OrderedMap) UnmarshalJSON(data []byte) error {
	// Create a temporary map for JSON unmarshaling
	tmpMap := make(map[string]interface{})
	if err := json.Unmarshal(data, &tmpMap); err != nil {
		return err
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	// Clear existing data without locking (we already have the lock)
	om.nodeMap = make(map[any]*Node)
	om.head = nil
	om.tail = nil
	om.length = 0

	// Add items to ordered map
	for k, v := range tmpMap {
		// Use internal set method to avoid double locking
		if err := om.set(k, v); err != nil {
			return err
		}
	}

	return nil
}

// internal set method without locking
func (om *OrderedMap) set(key, value any) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	if node, exists := om.nodeMap[key]; exists {
		node.Value = value
		return nil
	}

	newNode := &Node{
		Key:   key,
		Value: value,
	}

	if om.head == nil {
		om.head = newNode
		om.tail = newNode
	} else {
		newNode.prev = om.tail
		om.tail.next = newNode
		om.tail = newNode
	}

	om.nodeMap[key] = newNode
	om.length++
	return nil
}

// First returns the first key-value pair in the map.
// Returns nil values and false if the map is empty.
// This method is thread-safe.
//
// Example:
//
//	if key, value, exists := om.First(); exists {
//	    fmt.Printf("First element - Key: %v, Value: %v\n", key, value)
//	}
func (om *OrderedMap) First() (key, value any, exists bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if om.head == nil {
		return nil, nil, false
	}
	return om.head.Key, om.head.Value, true
}

// Last returns the last key-value pair in the map.
// Returns nil values and false if the map is empty.
// This method is thread-safe.
//
// Example:
//
//	if key, value, exists := om.Last(); exists {
//	    fmt.Printf("Last element - Key: %v, Value: %v\n", key, value)
//	}
func (om *OrderedMap) Last() (key, value any, exists bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if om.tail == nil {
		return nil, nil, false
	}
	return om.tail.Key, om.tail.Value, true
}

// Reverse returns a new OrderedMap with all elements in reverse order.
// This method is thread-safe.
//
// Example:
//
//	reversed := om.Reverse()
//	fmt.Println(reversed.String())
func (om *OrderedMap) Reverse() *OrderedMap {
	om.mu.RLock()
	defer om.mu.RUnlock()

	reversed := NewOrderedMap()
	current := om.tail
	for current != nil {
		_ = reversed.Set(current.Key, current.Value)
		current = current.prev
	}
	return reversed
}

// Filter returns a new OrderedMap containing only the elements that satisfy
// the given predicate function.
// This method is thread-safe.
//
// Example:
//
//	filtered := om.Filter(func(key, value any) bool {
//	    // Keep only values greater than 10
//	    if val, ok := value.(int); ok {
//	        return val > 10
//	    }
//	    return false
//	})
func (om *OrderedMap) Filter(predicate func(key, value any) bool) *OrderedMap {
	om.mu.RLock()
	defer om.mu.RUnlock()

	filtered := NewOrderedMap()
	current := om.head
	for current != nil {
		if predicate(current.Key, current.Value) {
			_ = filtered.Set(current.Key, current.Value)
		}
		current = current.next
	}
	return filtered
}

// Map creates a new OrderedMap by transforming each element using
// the given mapping function.
// This method is thread-safe.
//
// Example:
//
//	doubled := om.Map(func(key, value any) (any, any) {
//	    if val, ok := value.(int); ok {
//	        return key, val * 2
//	    }
//	    return key, value
//	})
func (om *OrderedMap) Map(mapper func(key, value any) (any, any)) *OrderedMap {
	om.mu.RLock()
	defer om.mu.RUnlock()

	mapped := NewOrderedMap()
	current := om.head
	for current != nil {
		newKey, newValue := mapper(current.Key, current.Value)
		_ = mapped.Set(newKey, newValue)
		current = current.next
	}
	return mapped
}

// JSONOptions represents configuration options for JSON marshaling/unmarshaling
type JSONOptions struct {
	// KeyAsString determines whether to force convert all keys to strings
	KeyAsString bool
	// PreserveType attempts to preserve the original type of numeric values
	PreserveType bool
	// PrettyPrint formats the JSON output with indentation
	PrettyPrint bool
}

// ToJSON converts the OrderedMap to a JSON byte array with the specified options.
// This method is thread-safe.
//
// Example:
//
//	opts := &JSONOptions{
//	    KeyAsString: true,
//	    PreserveType: true,
//	    PrettyPrint: true,
//	}
//	data, err := om.ToJSON(opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (om *OrderedMap) ToJSON(opts *JSONOptions) ([]byte, error) {
	if opts == nil {
		opts = &JSONOptions{
			KeyAsString:  true,
			PreserveType: false,
			PrettyPrint:  false,
		}
	}

	om.mu.RLock()
	defer om.mu.RUnlock()

	tmpMap := make(map[string]interface{})
	current := om.head
	for current != nil {
		var key string
		if opts.KeyAsString {
			key = fmt.Sprintf("%v", current.Key)
		} else if strKey, ok := current.Key.(string); ok {
			key = strKey
		} else {
			return nil, fmt.Errorf("non-string key %v cannot be converted to JSON", current.Key)
		}

		value := current.Value
		if opts.PreserveType {
			// Attempt to preserve numeric types
			if str, ok := value.(string); ok {
				if v, err := json.Number(str).Int64(); err == nil {
					value = v
				} else if v, err := json.Number(str).Float64(); err == nil {
					value = v
				}
			}
		}

		tmpMap[key] = value
		current = current.next
	}

	if opts.PrettyPrint {
		return json.MarshalIndent(tmpMap, "", "  ")
	}
	return json.Marshal(tmpMap)
}

// FromJSON populates the OrderedMap from a JSON byte array with the specified options.
// This method is thread-safe.
//
// Example:
//
//	opts := &JSONOptions{
//	    KeyAsString: true,
//	    PreserveType: true,
//	}
//	err := om.FromJSON(data, opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (om *OrderedMap) FromJSON(data []byte, opts *JSONOptions) error {
	if opts == nil {
		opts = &JSONOptions{
			KeyAsString:  true,
			PreserveType: false,
		}
	}

	var tmpMap map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(data))
	if opts.PreserveType {
		d.UseNumber()
	}
	if err := d.Decode(&tmpMap); err != nil {
		return err
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	// Clear existing data
	om.nodeMap = make(map[any]*Node)
	om.head = nil
	om.tail = nil
	om.length = 0

	// Add items to ordered map
	for k, v := range tmpMap {
		var key interface{} = k
		if !opts.KeyAsString {
			// Attempt to convert string key to appropriate type
			if i, err := strconv.ParseInt(k, 10, 64); err == nil {
				key = i
			} else if f, err := strconv.ParseFloat(k, 64); err == nil {
				key = f
			}
		}

		if opts.PreserveType {
			if num, ok := v.(json.Number); ok {
				if f, err := num.Float64(); err == nil {
					v = f
				}
			}
		}

		if err := om.set(key, v); err != nil {
			return err
		}
	}

	return nil
}
