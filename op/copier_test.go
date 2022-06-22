package op

import (
	"reflect"
	"testing"
	"time"
)

func TestNewCopier(t *testing.T) {
	src := "/testing/src"
	dest := "/testing/dest"
	since := time.Time{}
	until := time.Now()
	expect := Op{
		Identifier: "copy /testing/src",
		Runner: Copier{
			SourceDir: "/testing/",
			Filter:    "src",
			DestDir:   dest,
			Since:     since,
			Until:     until,
		},
	}
	copier := NewCopier(src, dest, since, until)

	if !reflect.DeepEqual(&expect, copier) {
		t.Errorf("unexpected copier field, expected=%#v, actual=%#v", expect, copier)
	}
}
