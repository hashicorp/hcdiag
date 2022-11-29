// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExampleTimeout_MarshalJSON provides some examples to demonstrate the usefulness of the Timeout
// struct as it relates to marshalling to JSON. Because results from diagnostic runs are written
// as JSON, this provides a more human-readable format than what the time.Duration type provides.
func ExampleTimeout_MarshalJSON() {
	var (
		b        []byte
		duration time.Duration
		timeout  Timeout
	)

	// When a time.Duration is marshalled to JSON, it renders as an integer,
	// representing the number of nanoseconds in the duration.
	duration = 3 * time.Second
	b, _ = json.Marshal(duration)
	fmt.Println(string(b)) // 3000000000

	// When a Timeout is marshalled to JSON, it renders in the same format
	// that Go uses to parse from strings into time.Duration, providing a
	// more human-readable output format.
	timeout = Timeout(duration)
	b, _ = json.Marshal(timeout)
	fmt.Println(string(b)) // "3s"

	// More complex time durations are also supported for human-readable output.
	timeout = Timeout((2 * time.Hour) + (31 * time.Minute) + (12 * time.Second))
	b, _ = json.Marshal(timeout)
	fmt.Println(string(b)) // "2h31m12s"

	// Output:
	// 3000000000
	// "3s"
	// "2h31m12s"
}
