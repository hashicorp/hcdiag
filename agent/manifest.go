// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/hcdiag/op"
)

// ManifestOp provides a subset of op state, specifically excluding results, so we can safely render metadata
// about ops without exposing customer info in manifest.json
type ManifestOp struct {
	ID       string    `json:"op"`
	Error    string    `json:"error"`
	Status   op.Status `json:"status"`
	Duration string    `json:"duration"`
}

func WalkResultsForManifest(results map[string]op.Op) []ManifestOp {
	m := make(map[string]any, 0)
	for k, v := range results {
		m[k] = any(v)
	}
	acc := make([]ManifestOp, 0)
	return walk(m, acc)
}

func walk(res map[string]any, acc []ManifestOp) []ManifestOp {
	for _, v := range res {
		switch o := v.(type) {
		case op.Op:
			manifestOp := ManifestOp{
				ID:       o.Identifier,
				Error:    o.ErrString,
				Status:   o.Status,
				Duration: fmt.Sprintf("%d", o.End.Sub(o.Start).Nanoseconds()),
			}
			acc = append(acc, manifestOp)
			spew.Dump("iterate", acc)
			walk(o.Result, acc)
		default:
			spew.Dump("return", acc)
			continue
		}
	}
	return acc
}
