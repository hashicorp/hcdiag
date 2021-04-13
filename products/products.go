package products

import (
	"fmt"

	s "github.com/hashicorp/host-diagnostics/seeker"
	"github.com/hashicorp/host-diagnostics/util"
)

const DebugSeconds = 3

// GetSeekers provides product Seekers for gathering info.
func GetSeekers(product string, tmpDir string) (seekers []*s.Seeker, err error) {
	if product == "" {
		return seekers, err
	} else if product == "nomad" {
		seekers = append(seekers, NomadSeekers(tmpDir)...)
	} else if product == "vault" {
		seekers = append(seekers, VaultSeekers(tmpDir)...)
	} else {
		err = fmt.Errorf("unsupported product '%s'", product)
	}
	return seekers, err
}

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
