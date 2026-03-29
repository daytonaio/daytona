// Simple verification that TTY field works in generated client libraries
package main

import (
	"encoding/json"
	"fmt"

	toolbox "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

func main() {
	fmt.Println("🔍 Verifying TTY implementation in client libraries...")

	// Test 1: Create ExecuteRequest with TTY field
	req := toolbox.NewExecuteRequest("echo 'Hello TTY'")
	req.SetTty(true)
	req.SetCwd("/tmp")
	req.SetTimeout(30)

	fmt.Println("✅ Created ExecuteRequest with TTY field")

	// Test 2: Verify TTY field getter/setter works
	if req.GetTty() != true {
		fmt.Println("❌ TTY field getter/setter failed")
		return
	}
	fmt.Println("✅ TTY field getter/setter works correctly")

	// Test 3: Verify JSON serialization includes TTY field
	data, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("❌ JSON marshaling failed: %v\n", err)
		return
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		fmt.Printf("❌ JSON unmarshaling failed: %v\n", err)
		return
	}

	if ttyValue, exists := jsonMap["tty"]; !exists || ttyValue != true {
		fmt.Printf("❌ TTY field not properly serialized in JSON: %v\n", jsonMap)
		return
	}
	fmt.Println("✅ TTY field properly serialized in JSON")

	// Test 4: Verify JSON deserialization works
	testJSON := `{"command": "test command", "tty": true, "cwd": "/home", "timeout": 60}`
	var deserializedReq toolbox.ExecuteRequest
	err = json.Unmarshal([]byte(testJSON), &deserializedReq)
	if err != nil {
		fmt.Printf("❌ JSON deserialization failed: %v\n", err)
		return
	}

	if !deserializedReq.GetTty() {
		fmt.Println("❌ TTY field not properly deserialized from JSON")
		return
	}
	fmt.Println("✅ TTY field properly deserialized from JSON")

	// Test 5: Test TTY field omission when false
	reqFalse := toolbox.NewExecuteRequest("echo 'no TTY'")
	// Don't set TTY (should default to false)

	dataFalse, err := json.Marshal(reqFalse)
	if err != nil {
		fmt.Printf("❌ JSON marshaling failed: %v\n", err)
		return
	}

	var jsonMapFalse map[string]interface{}
	err = json.Unmarshal(dataFalse, &jsonMapFalse)
	if err != nil {
		fmt.Printf("❌ JSON unmarshaling failed: %v\n", err)
		return
	}

	// TTY field should be omitted when not set (omitempty behavior)
	if _, exists := jsonMapFalse["tty"]; exists {
		fmt.Printf("❌ TTY field should be omitted when not set: %v\n", jsonMapFalse)
		return
	}
	fmt.Println("✅ TTY field properly omitted when not set")

	fmt.Println("")
	fmt.Println("🎉 All TTY implementation tests passed!")
	fmt.Println("📋 Summary:")
	fmt.Println("   • ExecuteRequest includes TTY field")
	fmt.Println("   • TTY field has proper getter/setter methods")
	fmt.Println("   • JSON serialization includes TTY when set")
	fmt.Println("   • JSON deserialization works correctly")
	fmt.Println("   • TTY field omitted from JSON when not set (omitempty)")
	fmt.Println("")
	fmt.Println("✨ The TTY flag implementation is ready for use!")
}
