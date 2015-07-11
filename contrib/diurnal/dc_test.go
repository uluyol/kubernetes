package main

import (
	"testing"
	"time"
)

func equalsTimeCounts(a, b []timeCount) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].time != b[i].time || a[i].count != b[i].count {
			return false
		}
	}
	return true
}

func TestParseTimeCounts(t *testing.T) {
	cases := []struct {
		times  string
		counts string
		out    []timeCount
		err    bool
	}{
		{
			"00:00:01Z,00:02Z,03:00Z,04:00Z", "1,4,1,8", []timeCount{
				{time.Second, 1},
				{2 * time.Minute, 4},
				{3 * time.Hour, 1},
				{4 * time.Hour, 8},
			}, false,
		},
		{"00:00Z,00:01Z", "1,0", []timeCount{{0, 1}, {1 * time.Minute, 0}}, false},
		{"00:01Z,00:02Z,00:05Z,00:02Z", "1,2,3,4", nil, true},
		{"00:00+00,00:01+00:00,01:00Z", "0,-1,0", nil, true},
		{"-00:01Z,01:00Z", "0,1", nil, true},
		{"00:00Z", "1,2,3", nil, true},
	}
	for i, test := range cases {
		out, err := parseTimeCounts(test.times, test.counts)
		if test.err && err == nil {
			t.Errorf("case %d: expected error", i)
		} else if !test.err && err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
		}
		if !test.err {
			if !equalsTimeCounts(test.out, out) {
				t.Errorf("case %d: expected timeCounts: %v got %v", i, test.out, out)
			}
		}
	}
}
