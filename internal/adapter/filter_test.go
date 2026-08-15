package adapter

import (
	"reflect"
	"testing"
)

func TestSplitFilter(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{`-k unit`, []string{"-k", "unit"}},
		{`-k "create or update"`, []string{"-k", "create or update"}},
		{`--testPathPattern='unit|smoke'`, []string{"--testPathPattern=unit|smoke"}},
		{``, []string{}},
		{`  -k   unit  `, []string{"-k", "unit"}},
		{`-k "unmatched`, []string{"-k", "unmatched"}},
	}
	for _, c := range cases {
		if got := SplitFilter(c.in); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("SplitFilter(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
