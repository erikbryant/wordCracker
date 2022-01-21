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

func TestSortUnique(t *testing.T) {
	testCases := []struct {
		w        []string
		expected []string
	}{
		{[]string{}, []string{}},
	}

	for _, testCase := range testCases {
		answer := sortUnique(testCase.w)
		if !equal(answer, testCase.expected) {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.w, testCase.expected, answer)
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

func TestSuggestGuess(t *testing.T) {
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
		answer := suggestGuess(testCase.m, testCase.g)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v %s expected %s, got %s", testCase.m, testCase.g, testCase.expected, answer)
		}
	}
}

// Masks to try
//
// audio toads about baton
// ybbby,yyybb,yyyby,ggggg

// Word #17 - shire
// gggbg,bbbyg,ggbyb,ybyyg,ybyyb,bgybb,bgbyg,bbbbg,bgbbg,bgbyg,ggbgg,byggb,byyyy,gbggg,bbbyb,bybby,bybyy,gggyy,bbbyy,bbbbb,bggyb
