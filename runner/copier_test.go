package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCopier(t *testing.T) {
	src := "/testing/src"
	dest := "/testing/dest"
	since := time.Time{}
	until := time.Now()
	expect := &Copier{
		SourceDir: "/testing/",
		Filter:    "src",
		DestDir:   dest,
		Since:     since,
		Until:     until,
	}
	copier := NewCopier(src, dest, since, until, nil)
	assert.Equal(t, expect, copier)
}
