package main

import (
	"encoding/json"
	"fmt"

	"github.com/mstgnz/orderedmap"
)

type Address struct {
	Street  string
	City    string
	Country string
}

type User struct {
	ID      int
	Name    string
	Email   string
	Address Address
}

func main() {
	// Create a new ordered map instance
	om := orderedmap.NewOrderedMap()

	// Create data with nested structure
	user1 := User{
		ID:    1,
		Name:  "Mesut Genez",
		Email: "mesutgenez@gmail.com",
		Address: Address{
			Street:  "123 Tech Street",
			City:    "Istanbul",
			Country: "Turkey",
		},
	}

	user2 := User{
		ID:    2,
		Name:  "John Doe",
		Email: "john@example.com",
		Address: Address{
			Street:  "456 Code Avenue",
			City:    "New York",
			Country: "USA",
		},
	}

	// Add data to OrderedMap
	om.Set("user1", user1)
	om.Set("user2", user2)

	// Convert OrderedMap to JSON
	jsonData, err := json.MarshalIndent(om, "", "    ")
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return
	}

	fmt.Println("Marshalled JSON:")
	fmt.Println(string(jsonData))

	// Convert JSON to OrderedMap
	jsonStr := `{
		"user3": {
			"ID": 3,
			"Name": "Alice Smith",
			"Email": "alice@example.com",
			"Address": {
				"Street": "789 Dev Road",
				"City": "London",
				"Country": "UK"
			}
		}
	}`

	newMap := orderedmap.NewOrderedMap()
	err = json.Unmarshal([]byte(jsonStr), newMap)
	if err != nil {
		fmt.Printf("Unmarshal error: %v\n", err)
		return
	}

	// Check unmarshaled data
	if user3, exists := newMap.Get("user3"); exists {
		fmt.Println("\nUnmarshalled Data:")
		fmt.Printf("User3: %+v\n", user3)
	}

	// List all users
	fmt.Println("\nAll Users:")
	om.Range(func(key, value any) bool {
		fmt.Printf("%v: %+v\n", key, value)
		return true
	})
}
