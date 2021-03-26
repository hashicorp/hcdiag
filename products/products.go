package products

import (
	"fmt"
)

// ProductCommand struct
type ProductCommand struct {
	Attribute string
	Command   string
	Arguments []string
	Format    string
}

// ProductCommands stuff
func ProductCommands(productName string) []ProductCommand {
	ProductCommands := make([]ProductCommand, 0)

	switch {
	case productName == "terraform":
		ProductCommands = append(ProductCommands,
			ProductCommand{
				Attribute: "example",
				Command:   "terraform",
				Arguments: []string{"version"},
			})

	case productName == "vault":
		ProductCommands = append(ProductCommands,
			ProductCommand{
				Attribute: "vault status -format json",
				Command:   "vault",
				Arguments: []string{"status", "-format=json"},
				Format:    "json",
			},
			ProductCommand{
				Attribute: "vault version",
				Command:   "vault",
				Arguments: []string{"version"},
				Format:    "string",
			},
			ProductCommand{
				Attribute: "vault read sys/health",
				Command:   "vault",
				Arguments: []string{"read", "sys/health", "-format=json"},
				Format:    "json",
			},
			ProductCommand{
				Attribute: "vault read sys/host-info",
				Command:   "vault",
				Arguments: []string{"read", "sys/host-info", "-format=json"},
				Format:    "json",
			},
			ProductCommand{
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
