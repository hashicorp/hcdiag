package seeker

// Validation callback functions for double-checking seeker results

import (
	"fmt"
	"strings"
)

// ResultNotContains' returned func will error if the Seeker's Result contains a given substring.
func ResultNotContains(s *Seeker, substr string, err error) func() error {
	return func() error {
		result, ok := s.Result.(string)

		if !ok { // will happen if Result is nil (i.e. Run() has not run yet)
			return fmt.Errorf("not a string: %#v", result)
		}

		if strings.Contains(result, substr) {
			// return the error that we were given initially
			return err
		}

		return nil
	}
}
