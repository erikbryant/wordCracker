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
	}

	for _, testCase := range testCases {
		answer := matchSingleWord(testCase.w, testCase.mask, testCase.c)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %s/%s/%s expected %t, got %t", testCase.w, testCase.mask, testCase.c, testCase.expected, answer)
		}
	}
}

func TestInitPossibleLetters(t *testing.T) {
	testCases := []struct {
		word     string
		expected []map[byte]bool
	}{
		{"", []map[byte]bool{}},
		{"a", []map[byte]bool{map[byte]bool{97: true}}},
		{"bf", []map[byte]bool{map[byte]bool{98: true, 102: true}, map[byte]bool{98: true, 102: true}}},
	}

	for _, testCase := range testCases {
		answer := initPossibleLetters(testCase.word)
		if !equalMap(answer, testCase.expected) {
			t.Errorf("ERROR: For '%s' expected %v, got %v", testCase.word, testCase.expected, answer)
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
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg,byggb
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg,byggb,byyyy
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg,byggb,byyyy,gbggg,bbbyb,bybby,bybyy,gggyy
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg,byggb,byyyy,gbggg,bbbyb,bybby,bybyy,gggyy,bbbyy,bggyb
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg,byggb,byyyy,gbggg,bbbyb,bybby,bybyy,gggyy,bbbyy,bbbbb,bggyb
