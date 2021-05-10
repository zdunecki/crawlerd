package test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Diff(t *testing.T, title string, vExpect, vCurrent interface{}) {
	if diff := cmp.Diff(vExpect, vCurrent); diff != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", title, diff)
	}
}
