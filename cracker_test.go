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

func equalInt(a, b []int) bool {
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

func equalInt2(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !equalInt(a[i], b[i]) {
			return false
		}
	}

	return true
}

func equalByte(a, b []byte) bool {
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

func equalScore(a, b score) bool {
	return a.score == b.score && a.word == b.word
}

func TestValidMask(t *testing.T) {
	testCases := []struct {
		m        string
		len      int
		expected bool
	}{
		{"", 0, true},
		{"b", 1, true},
		{"y", 1, true},
		{"g", 1, true},
		{"x", 1, false},
		{"bbyyg", 5, true},
		{"bbyyg", 3, false},
	}

	for _, testCase := range testCases {
		answer, err := validMask(testCase.m, testCase.len)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %s %d expected %t, got %t", testCase.m, testCase.len, testCase.expected, answer)
		}
		if answer && err != nil {
			t.Errorf("ERROR: For %s %d expected no error, got an error", testCase.m, testCase.len)
		}
		if !answer && err == nil {
			t.Errorf("ERROR: For %s %d expected an error, got nil", testCase.m, testCase.len)
		}
	}
}

func TestUnpackMasks(t *testing.T) {
	testCases := []struct {
		m           string
		expected    []string
		expectError bool
	}{
		{"", []string{""}, false},
		{"bbbyy", []string{"bbbyy"}, false},
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

func TestUnpackGuessed(t *testing.T) {
	testCases := []struct {
		g           string
		expected1   []string
		expected2   []string
		expectError bool
	}{
		{"", []string{}, []string{}, true},
		{"bbbyy", []string{}, []string{}, true},
		{"bbbyy,gy,gyybb,gbygb,ggbgg", []string{}, []string{}, true},
		{"bbbyy,gybbbbb,gyybb,gbygb,ggbgg", []string{}, []string{}, true},
		{"bbbyy,gybbb,gyybb,gbygb,ggbgg,asdff", []string{}, []string{}, true},
		{"moist/bbbyy,house/gybbb,walks/gyybb", []string{"moist", "house", "walks"}, []string{"bbbyy", "gybbb", "gyybb"}, false},
		{"to/bb,of/gy,is/gg", []string{"to", "of", "is"}, []string{"bb", "gy", "gg"}, false},
	}

	for _, testCase := range testCases {
		answer1, answer2, err := unpackGuessed(testCase.g)
		if !equal(answer1, testCase.expected1) {
			t.Errorf("ERROR: For '%s' expected %v %v, got %v %v", testCase.g, testCase.expected1, testCase.expected2, answer1, answer2)
		}
		if !equal(answer2, testCase.expected2) {
			t.Errorf("ERROR: For '%s' expected %v %v, got %v %v", testCase.g, testCase.expected1, testCase.expected2, answer1, answer2)
		}
		if testCase.expectError && err == nil {
			t.Errorf("ERROR: For '%s' expected error:<something>, got error:%v", testCase.g, err)
		}
		if !testCase.expectError && err != nil {
			t.Errorf("ERROR: For '%s' expected error:nil, got error:%v", testCase.g, err)
		}
	}
}

func TestReplace(t *testing.T) {
	testCases := []struct {
		w        []byte
		a        byte
		b        byte
		expected []byte
	}{
		{[]byte{}, ' ', '_', []byte{}},
		{[]byte{'a', 'b'}, '_', 'A', []byte{'a', 'b'}},
		{[]byte{'a', 'b'}, 'a', 'c', []byte{'c', 'b'}},
		{[]byte{'a', 'b'}, 'b', '4', []byte{'a', '4'}},
	}

	for _, testCase := range testCases {
		answer := make([]byte, len(testCase.w))
		copy(answer, testCase.w)
		replace(answer, testCase.a, testCase.b)
		if !equalByte(answer, testCase.expected) {
			t.Errorf("ERROR: For %v %c %c expected %v, got %v", testCase.w, testCase.a, testCase.b, testCase.expected, answer)
		}
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		w        []byte
		b        byte
		expected bool
	}{
		{[]byte{}, ' ', false},
		{[]byte{'a', 'b'}, '_', false},
		{[]byte{'a', 'b'}, 'a', true},
		{[]byte{'a', 'b'}, 'b', true},
	}

	for _, testCase := range testCases {
		answer := contains(testCase.w, testCase.b)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v %c expected %t, got %t", testCase.w, testCase.b, testCase.expected, answer)
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
		{"those", "gbygb", "tress", true},
		{"those", "gbygb", "trest", true},
	}

	for _, testCase := range testCases {
		answer := matchSingleWord(testCase.w, testCase.mask, testCase.c)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %s/%s/%s expected %t, got %t", testCase.w, testCase.mask, testCase.c, testCase.expected, answer)
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

func TestApplyMasks(t *testing.T) {
	testCases := []struct {
		m        []string
		g        []string
		masks    []string
		expected []string
	}{
		{[]string{}, []string{}, []string{}, []string{}},
		{[]string{"foo"}, []string{"foo"}, []string{"ggg"}, []string{"foo"}},
		{[]string{"foo"}, []string{"foo"}, []string{"bbb"}, []string{}},
		{[]string{"foo"}, []string{"bar"}, []string{"ggg"}, []string{}},
	}

	for _, testCase := range testCases {
		answer := applyMasks(testCase.m, testCase.g, testCase.masks)
		if !equal(answer, testCase.expected) {
			t.Errorf("ERROR: For %v/%v/%v expected %v, got %v", testCase.m, testCase.g, testCase.masks, testCase.expected, answer)
		}
	}
}

func TestMakeMask(t *testing.T) {
	testCases := []struct {
		w        string
		g        string
		expected string
	}{
		{"", "", ""},
		{"abc", "ddd", "bbb"},
		{"abc", "aaa", "gbb"},
		{"abc", "cab", "yyy"},
		{"apple", "house", "bbbbg"},
	}

	for _, testCase := range testCases {
		answer := makeMask(testCase.w, testCase.g)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %s %s expected %s, got %s", testCase.w, testCase.g, testCase.expected, answer)
		}
	}
}

func TestFindMaxScore(t *testing.T) {
	testCases := []struct {
		s        []score
		g        string
		expected score
	}{
		{[]score{}, "abc", score{-1, ""}},
		{[]score{score{2, "aaa"}}, "abc", score{2, "aaa"}},
		{[]score{score{2, "abc"}}, "abc", score{-1, ""}},
		{[]score{score{2, "aaa"}, score{5, "abc"}}, "abc", score{2, "aaa"}},
	}

	for _, testCase := range testCases {
		answer := findMaxScore(testCase.s, testCase.g)
		if !equalScore(answer, testCase.expected) {
			t.Errorf("ERROR: For %v %s expected %v, got %v", testCase.s, testCase.g, testCase.expected, answer)
		}
	}
}

func TestSuggestGuessLetterrFreq(t *testing.T) {
	testCases := []struct {
		m        []string
		g        string
		expected string
	}{
		{[]string{""}, "", ""},
		{[]string{"abc"}, "", "abc"},
		{[]string{"abc"}, "abc", ""},
		{[]string{"abc", "def"}, "abc", "def"},
		{[]string{"abc", "def"}, "def", "abc"},
	}

	for _, testCase := range testCases {
		answer := suggestGuessLetterFreq(testCase.m, testCase.g)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v %s expected %s, got %s", testCase.m, testCase.g, testCase.expected, answer)
		}
	}
}

func TestPruneGuessables(t *testing.T) {
	testCases := []struct {
		g        []string
		w        string
		m        string
		expected []string
	}{
		// Degenerate cases
		{[]string{}, "", "", []string{}},
		{[]string{""}, "", "", []string{""}},

		// g
		{[]string{"c", "d"}, "c", "g", []string{"c"}},
		{[]string{"cat", "dog"}, "nap", "bgb", []string{"cat"}},
		{[]string{"cat", "dog"}, "tap", "bgb", []string{}},
		{[]string{"cat", "dog", "zaa"}, "nap", "bgb", []string{"cat", "zaa"}},

		// simple b
		{[]string{"c", "d"}, "c", "b", []string{"d"}},
		{[]string{"cat", "dog"}, "cry", "bbb", []string{"dog"}},

		// simple y
		{[]string{"cx", "dx"}, "ac", "by", []string{"cx"}},

		// b with g
		{[]string{"cat", "dog"}, "ccy", "gbb", []string{"cat"}},

		// b with y
		{[]string{"cat", "dog"}, "ycc", "byb", []string{"cat"}},

		// b with y and g
		{[]string{"cat", "dog"}, "dgg", "gyb", []string{"dog"}},
	}

	for _, testCase := range testCases {
		answer := pruneGuessables(testCase.g, testCase.w, testCase.m)
		if !equal(answer, testCase.expected) {
			t.Errorf("ERROR: For %v '%s' '%s' expected %v, got %v", testCase.g, testCase.w, testCase.m, testCase.expected, answer)
		}
	}
}

// Masks to try
//
// audio toads about baton
// ybbby,yyybb,yyyby,ggggg

// Word #17 - shire
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg,byggb,byyyy,gbggg,bbbyb,bybby,bybyy,gggyy,bbbyy,bbbbb,bggyb
