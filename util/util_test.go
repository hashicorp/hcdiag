package util

import (
	"testing"
)

func TestExecuteCommands(t *testing.T) {
	// Create list of command struct to execute
	testCommandStruct := []CommandStruct{{
		Attribute: "test",
		Command:   "echo",
		Arguments: []string{"test"},
		Format:    "string",
	}}

	// Craft expected result for comparison
	expectedResult := map[string]interface{}{
		"test": "test",
	}

	// Execute command and validate no error
	result, err := ExecuteCommands(testCommandStruct, false)
	if err != nil {
		t.Errorf("TestExecuteCommands failed: %s", err)
	}

	// Validate expected result
	if result["test"] != expectedResult["test"] {
		t.Errorf("Unexpected result: %s", result)
	}
}

// TarGz(sourceDir string, destFileName string) error

// InterfaceToJSON(mapVar map[string]interface{}) ([]byte, error)

// JSONToFile(JSON []byte, outFile string) error
