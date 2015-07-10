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
			"00:00:01,00:02,03:00,04:00", "1,4,1,8", []timeCount{
				{time.Second, 1},
				{2 * time.Minute, 4},
				{3 * time.Hour, 1},
				{4 * time.Hour, 8},
			}, false,
		},
		{"00:00,00:01", "1,0", []timeCount{{0, 1}, {1 * time.Minute, 0}}, false},
		{"00:01,00:02,00:05,00:02", "1,2,3,4", nil, true},
		{"00:00,00:01,01:00", "0,-1,0", nil, true},
		{"-00:01,01:00", "0,1", nil, true},
		{"00:00", "1,2,3", nil, true},
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

func TestSmoothChangesSimple(t *testing.T) {
	cases := []struct {
		tc          []timeCount
		rate        float64
		err         bool
		expectedLen int
	}{
		{[]timeCount{{1 * time.Second, 1}, {2 * time.Second, 20}}, 10, true, 0},
		{[]timeCount{{0, 0}, {1 * time.Hour, 60}}, 1, false, 120},
		{[]timeCount{{0, 0}, {1 * time.Hour, 60}}, 0.9, true, 0},
		{[]timeCount{{0, 60}, {1 * time.Hour, 0}}, 1, false, 120},
		{[]timeCount{{0, 0}, {1 * time.Minute, 4}, {3 * time.Minute, 12}}, 4, false, 6},
		{[]timeCount{{0, 0}, {1 * time.Minute, 4}, {3 * time.Minute, 13}}, 4, true, 0},
		{[]timeCount{{0, 0}, {1 * time.Minute, 4}, {3 * time.Minute, 11}}, 4, false, 6},
		{[]timeCount{{0, 0}, {1 * time.Minute, 4}, {3 * time.Minute, 0}}, 4, false, 3},
		{[]timeCount{{0, 0}, {1 * time.Minute, 4}, {3 * time.Minute, 8}}, 4, false, 4},
		{[]timeCount{{0, 0}, {1 * time.Minute, 4}, {3 * time.Minute, 9}}, 4, false, 6},
		{[]timeCount{{0, 0}, {2 * time.Minute, 1}}, 0.5, false, 2},
		{[]timeCount{{0, 0}, {4 * time.Minute, 2}}, 0.5, false, 4},
		{
			[]timeCount{
				{0, 0},
				{1 * time.Minute, 4},
				{2 * time.Minute, 9},
				{3 * time.Minute, 5},
				{4 * time.Minute, 0},
			}, 5, false, 5,
		},
	}
	for i, test := range cases {
		smoothed, err := smoothChanges(test.tc, test.rate)
		if test.err {
			if err == nil {
				t.Errorf("case %d: expected error", i)
			}
			continue
		}
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
		}
		if len(smoothed) < len(test.tc) {
			t.Errorf("case %d: smoothed list is too short, smoothed: %v", i, smoothed)
		}
		if len(smoothed) != test.expectedLen {
			t.Errorf("case %d: smoothed length: %d expected: %d", i, len(smoothed), test.expectedLen)
		}
		for j := 1; j < len(smoothed); j++ {
			if smoothed[j].time <= smoothed[j-1].time {
				t.Errorf("case %d: smoothed times not in strictly increasing order: %v", i, smoothed)
				break
			}
		}
	}
}
