package debug

import "fmt"

// filterArgs returns a string that contains one '-flagname=filter' pair for each filters element
func filterArgs(flagname string, filters []string) string {
	var arguments string
	for _, f := range filters {
		arguments = fmt.Sprintf("%s -%s=%s", arguments, flagname, f)
	}
	return arguments
}
