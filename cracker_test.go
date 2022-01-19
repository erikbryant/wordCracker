package main

import (
	"testing"
)

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func equalMap(a, b []map[byte]bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		aMap := a[i]
		bMap := b[i]

		if len(aMap) != len(bMap) {
			return false
		}
		for key := range aMap {
			if aMap[key] != bMap[key] {
				return false
			}
		}
	}

	return true
}

func TestUnpackMasks(t *testing.T) {
	testCases := []struct {
		m           string
		expected    []string
		expectError bool
	}{
		{"", []string{}, true},
		{"bbbyy", []string{}, true},
		{"bbbyy,gy,gyybb,gbygb,ggbgg", []string{}, true},
		{"bbbyy,gybbbbb,gyybb,gbygb,ggbgg", []string{}, true},
		{"bbbyy,gybbb,gyybb,gbygb,ggbgg,asdff", []string{}, true},
		{"bbbyy,gybbb,gyybb,gbygb,ggbgg", []string{"bbbyy", "gbygb", "ggbgg", "gybbb", "gyybb"}, false},
		{"bb,gy,gg", []string{"bb", "gg", "gy"}, false},
	}

	for _, testCase := range testCases {
		answer, err := unpackMasks(testCase.m)
		if !equal(answer, testCase.expected) {
			t.Errorf("ERROR: For '%s' expected %v, got %v", testCase.m, testCase.expected, answer)
		}
		if testCase.expectError && err == nil {
			t.Errorf("ERROR: For '%s' expected error:<something>, got error:%v", testCase.m, err)
		}
		if !testCase.expectError && err != nil {
			t.Errorf("ERROR: For '%s' expected error:nil, got error:%v", testCase.m, err)
		}
	}
}

func TestMatchSingleWord(t *testing.T) {
	testCases := []struct {
		w        string
		mask     string
		c        string
		expected bool
	}{
		{"pshaw", "bgggg", "tshaw", true},
		{"pshaw", "bgggg", "pshaw", false},
		{"pshaw", "bbbyy", "xxxxw", false},
		{"pshaw", "bbbyy", "xxxax", false},
		{"pshaw", "bbbyy", "xxxaw", false},
		{"third", "bbbyy", "ardor", false},
		{"pshaw", "bbbyy", "waxxx", false},
		{"pshaw", "bbbyy", "awxxx", false},
		{"pshaw", "bbbyy", "xxxwa", true},
		{"pshaw", "ggggg", "pshax", false},
		{"pshaw", "ggggg", "pshaw", true},
		{"pshaw", "gbbbg", "pxxxw", true},
		{"pshaw", "ggyyg", "psahw", true},
		{"alpha", "gggyb", "alpax", true},
		{"alpha", "bbbbb", "xyzzy", true},
		{"pha", "bbb", "zzy", true},
		{"pha", "bby", "zzp", true},
	}

	for _, testCase := range testCases {
		answer := matchSingleWord(testCase.w, testCase.mask, testCase.c)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %s/%s/%s expected %t, got %t", testCase.w, testCase.mask, testCase.c, testCase.expected, answer)
		}
	}
}

func TestFilterByLen(t *testing.T) {
	testCases := []struct {
		words    []string
		len      int
		expected []string
	}{
		{[]string{"psh", "gg", "w"}, 0, []string{}},
		{[]string{"psh", "gg", "w"}, 1, []string{"w"}},
		{[]string{"psh", "gg", "w"}, 2, []string{"gg"}},
		{[]string{"psh", "gg", "w"}, 3, []string{"psh"}},
		{[]string{"psh", "gg", "w"}, 4, []string{}},
	}

	for _, testCase := range testCases {
		answer := filterByLen(testCase.words, testCase.len)
		if !equal(answer, testCase.expected) {
			t.Errorf("ERROR: For %v %d expected %v, got %v", testCase.words, testCase.len, testCase.expected, answer)
		}
	}
}

func TestMatchMasks(t *testing.T) {
	testCases := []struct {
		w        string
		mask     []string
		c        []string
		expected bool
	}{
		{"pshaw", []string{"bgggg"}, []string{"tshaw"}, true},
		{"pshaw", []string{"bgggg", "gbggg"}, []string{"tshaw"}, false},
		{"shaw", []string{"bggg"}, []string{"thaw"}, true},
		{"shaw", []string{"bggg", "gbgg"}, []string{"thaw"}, false},
	}

	for _, testCase := range testCases {
		answer := matchMasks(testCase.w, testCase.mask, testCase.c)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %s/%v/%v expected %t, got %t", testCase.w, testCase.mask, testCase.c, testCase.expected, answer)
		}
	}
}

// Masks to try
//
// audio toads about baton
// ybbby,yyybb,yyyby,ggggg

// Word #17 - shire
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg,byggb,byyyy,gbggg,bbbyb,bybby,bybyy,gggyy,bbbyy,bbbbb,bggyb
