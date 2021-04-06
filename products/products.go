package products

import (
	"fmt"

	"github.com/hashicorp/host-diagnostics/util"
)

// ProductCommands stuff
func ProductCommands(productName string, tempDir string) []util.CommandStruct {
	var ProductCommands []util.CommandStruct

	switch {
	case productName == "terraform":
		ProductCommands = append(ProductCommands,
			util.CommandStruct{
				Attribute: "example",
				Command:   "terraform",
				Arguments: []string{"version"},
			})

	case productName == "nomad":
		ProductCommands = append(ProductCommands, NomadCommands(tempDir)...)

	case productName == "vault":
		ProductCommands = append(ProductCommands, VaultCommands(tempDir)...)

	default:
		fmt.Println("default")

	}

	return ProductCommands
}
