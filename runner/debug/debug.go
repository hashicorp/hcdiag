package debug

import (
	"fmt"
	"math/rand"
)

const alphaNumBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ012345689"

// productFilterString takes a product name and a slice of filter strings, and produces valid, product-specific filter flags.
// The returned string is in the form " -target=metrics -target=pprof" (for Vault), " -capture=host" (for Consul), or " -event-topic=Allocation" (for Nomad)
func productFilterString(product string, filters []string) (string, error) {
	var filterString string
	var legalFilters map[string]bool
	var optFlag string

	// Define valid filter flagnames and values for all products
	nomadOptFlag := "event-topic"
	nomadFilters := map[string]bool{
		"*":          true,
		"ACLToken":   true,
		"ACLPolicy":  true,
		"ACLRole":    true,
		"Job":        true,
		"Allocation": true,
		"Deployment": true,
		"Evaluation": true,
		"Node":       true,
		"Service":    true,
	}

	vaultOptFlag := "target"
	vaultFilters := map[string]bool{
		"config":             true,
		"host":               true,
		"metrics":            true,
		"pprof":              true,
		"replication-status": true,
		"server-status":      true,
	}

	consulOptFlag := "capture"
	consulFilters := map[string]bool{
		"agent":   true,
		"host":    true,
		"members": true,
		"metrics": true,
		"logs":    true,
		"pprof":   true,
	}

	switch product {
	case "nomad":
		legalFilters = nomadFilters
		optFlag = nomadOptFlag
	case "vault":
		legalFilters = vaultFilters
		optFlag = vaultOptFlag
	case "consul":
		legalFilters = consulFilters
		optFlag = consulOptFlag
	default:
		return "", fmt.Errorf("invalid product used in debug.productFilterString(): %s", product)
	}

	for _, f := range filters {
		if !legalFilters[f] {
			return "", fmt.Errorf("invalid filter string (%s) for %s used in debug.productFilterString()", f, product)
		}
		// includes leading space
		filterString = fmt.Sprintf("%s -%s=%s", filterString, optFlag, f)
	}

	return filterString, nil
}

// randAlphanumString takes an integer n, and returns a pseudorandom string of len(n) composed of the characters (a-z A-Z 0-9).
func randAlphanumString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphaNumBytes[rand.Intn(len(alphaNumBytes))]
	}
	return string(b)
}
