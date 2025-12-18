// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/hcdiag/op"
)

func TestWalkResultsForManifest(t *testing.T) {
	testTable := []struct {
		desc          string
		ops           map[string]op.Op
		expectedCount int
	}{
		{
			desc:          "Empty map produces empty slice of ManifestOp",
			ops:           map[string]op.Op{},
			expectedCount: 0,
		},
		{
			desc: "Simple result value types are extracted",
			ops: map[string]op.Op{
				"opname": {
					Result: map[string]any{
						"result1": "string result",
					},
				},
			},
			expectedCount: 1,
		},
		{
			desc: "Shallow nested result value types are extracted",
			ops: map[string]op.Op{
				// Outer Op should be counted
				"do host": {
					Result: map[string]any{
						// Inner Op Level 1 should be counted
						"/etc/hosts": op.Op{
							Result: map[string]any{
								"shell": "##\nHost Database...",
							},
						},
						// Inner Op Level 1 should be counted
						"memory": op.Op{
							Result: map[string]any{
								"memoryInfo": "Memory Info",
							},
						},
					},
				},
			},
			expectedCount: 3,
		},
		{
			desc: "Deeply nested result value types are extracted",
			ops: map[string]op.Op{
				// Outer Op should be counted
				"do host": {
					Result: map[string]any{
						// Inner Op Level 1 should be counted
						"level1InnerOp1": op.Op{
							Result: map[string]any{
								// Inner Op Level 2 should be counted
								"level2InnerOp1": op.Op{
									Result: map[string]any{
										// Inner Op Level 3 should be counted
										"level3InnerOp1": op.Op{
											Result: map[string]any{
												"result": "value",
											},
										},
										"result": "value",
									},
								},
								// Inner Op Level 2 should be counted
								"level2InnerOp2": op.Op{
									Result: map[string]any{
										"result": "value",
									},
								},
							},
						},
					},
				},
			},
			expectedCount: 5,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.desc, func(t *testing.T) {
			manifestOps := WalkResultsForManifest(tc.ops)
			assert.Equal(t, tc.expectedCount, len(manifestOps))
		})
	}
}
