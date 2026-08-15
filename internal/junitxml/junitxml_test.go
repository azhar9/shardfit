package junitxml

import (
	"testing"
)

const nested = `<?xml version="1.0" encoding="utf-8"?>
<testsuites>
  <testsuite name="outer" tests="2">
    <testcase classname="A" name="one" time="1.234"/>
    <testsuite name="inner" tests="1">
      <testcase classname="B" name="two" time="0.5"/>
    </testsuite>
  </testsuite>
</testsuites>`

const bare = `<testsuite name="single" tests="1">
  <testcase classname="C" name="three" time="2"/>
</testsuite>`

func TestParseNestedSuites(t *testing.T) {
	cases, err := Parse([]byte(nested))
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 2 {
		t.Fatalf("cases = %d, want 2", len(cases))
	}
	if cases[0].Name != "one" || cases[1].Name != "two" {
		t.Fatalf("cases = %+v", cases)
	}
}

func TestParseBareSuite(t *testing.T) {
	cases, err := Parse([]byte(bare))
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 1 || cases[0].Name != "three" {
		t.Fatalf("cases = %+v", cases)
	}
}

func TestParseGarbage(t *testing.T) {
	if _, err := Parse([]byte("not xml")); err == nil {
		t.Fatal("want error for garbage")
	}
}

func TestTimeMs(t *testing.T) {
	if got := (Case{Time: 1.234}).TimeMs(); got != 1234 {
		t.Fatalf("TimeMs = %d, want 1234", got)
	}
}

func TestSumByID(t *testing.T) {
	cases := []Case{
		{Classname: "A", Name: "x", Time: 0.1},
		{Classname: "A", Name: "x", Time: 0.05}, // retry: summed
		{Classname: "A", Name: "y", Time: 0.2},
	}
	got := SumByID(cases, func(c Case) string { return c.Classname + "." + c.Name })
	if got["A.x"] != 150 || got["A.y"] != 200 {
		t.Fatalf("SumByID = %v, want A.x=150 A.y=200", got)
	}
}

func TestParseSkipsUnknownElements(t *testing.T) {
	doc := `<testsuite name="s"><properties><property name="p" value="v"/></properties>
  <testcase classname="A" name="x" time="1"/></testsuite>`
	cases, err := Parse([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 1 {
		t.Fatalf("cases = %d, want 1", len(cases))
	}
}
