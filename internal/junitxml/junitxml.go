// Package junitxml parses the JUnit XML reports that test runners already
// emit — the duration transport between buckets and the timing store.
package junitxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
)

// Case is one testcase element.
type Case struct {
	Classname string  `xml:"classname,attr"`
	Name      string  `xml:"name,attr"`
	Time      float64 `xml:"time,attr"`
}

type suites struct {
	Suites []suite `xml:"testsuite"`
}

type suite struct {
	Cases  []Case  `xml:"testcase"`
	Suites []suite `xml:"testsuite"`
}

// Parse accepts a <testsuites> or bare <testsuite> root and returns all
// testcases, including those in nested suites. Unknown elements are ignored.
func Parse(data []byte) ([]Case, error) {
	if bytes.Contains(data, []byte("<testsuites")) {
		var s suites
		if err := xml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("parse JUnit XML: %w", err)
		}
		return flatten(s.Suites), nil
	}
	var s suite
	if err := xml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse JUnit XML: %w", err)
	}
	return flatten([]suite{s}), nil
}

func flatten(ss []suite) []Case {
	var out []Case
	for _, s := range ss {
		out = append(out, s.Cases...)
		out = append(out, flatten(s.Suites)...)
	}
	return out
}

// TimeMs converts the time attribute (seconds) to milliseconds.
func (c Case) TimeMs() int64 {
	return int64(math.Round(c.Time * 1000))
}

// SumByID maps cases to durations (ms) via idFn, summing cases that share
// an id (retries, parametrized runs). Cases with an empty id are skipped.
func SumByID(cases []Case, idFn func(Case) string) map[string]int64 {
	out := map[string]int64{}
	for _, c := range cases {
		id := idFn(c)
		if id == "" {
			continue
		}
		out[id] += c.TimeMs()
	}
	return out
}
