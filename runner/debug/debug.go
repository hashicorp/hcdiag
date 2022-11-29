// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package debug

import (
	"fmt"
	"strings"
)

// filterArgs returns a string that contains one '-flagname=filter' pair for each element in filters
func filterArgs(flagname string, filters []string) string {
	var arguments strings.Builder
	for _, f := range filters {
		_, _ = fmt.Fprintf(&arguments, " -%s=%s", flagname, f)
	}
	return arguments.String()
}
