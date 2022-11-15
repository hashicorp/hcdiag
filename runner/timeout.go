package runner

import (
	"fmt"
	"time"
)

// Timeout is a time.Duration, which allows for custom JSON marshalling. When marshalled, it will be converted into
// a duration string, rather than an integer representing nanoseconds. For example, 3000000000 becomes "3s".
type Timeout time.Duration

func (t Timeout) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, time.Duration(t).String())), nil
}
