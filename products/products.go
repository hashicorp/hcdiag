package products

import (
	"fmt"

	"github.com/hashicorp/host-diagnostics/util"
)

// ProductCommands stuff
func ProductCommands(productName string) []util.CommandStruct {
	ProductCommands := make([]util.CommandStruct, 0)

	switch {
	case productName == "terraform":
		ProductCommands = append(ProductCommands,
			util.CommandStruct{
				Attribute: "example",
				Command:   "terraform",
				Arguments: []string{"version"},
			})

	case productName == "vault":
		ProductCommands = append(ProductCommands,
			util.CommandStruct{
				Attribute: "vault status -format json",
				Command:   "vault",
				Arguments: []string{"status", "-format=json"},
				Format:    "json",
			},
			util.CommandStruct{
				Attribute: "vault version",
				Command:   "vault",
				Arguments: []string{"version"},
				Format:    "string",
			},
			util.CommandStruct{
				Attribute: "vault read sys/health",
				Command:   "vault",
				Arguments: []string{"read", "sys/health", "-format=json"},
				Format:    "json",
			},
			util.CommandStruct{
				Attribute: "vault read sys/host-info",
				Command:   "vault",
				Arguments: []string{"read", "sys/host-info", "-format=json"},
				Format:    "json",
			},
			util.CommandStruct{
				Attribute: "vault read sys/seal-status",
				Command:   "vault",
				Arguments: []string{"read", "sys/seal-status", "-format=json"},
				Format:    "json",
			})

	default:
		fmt.Println("default")

	}

	return ProductCommands
}
