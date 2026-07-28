package main

import (
	"reflect"
	"testing"
)

func TestSplitList(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"com", []string{"com"}},
		{"com,net, org ", []string{"com", "net", "org"}},
		{" , ,", nil},
		{"homoglyph,tld-swap", []string{"homoglyph", "tld-swap"}},
	}
	for _, c := range cases {
		if got := splitList(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitList(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
