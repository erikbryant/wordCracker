package main

// go fmt && golint && go test && go run cracker.go -cpuprofile cpu.prof && echo top | go tool pprof cpu.prof

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"sort"
	"strings"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	masks      = flag.String("masks", "bbbyy,gybbb,gyybb,gbygb,ggbgg", "masks from previous games in the form of yyybb,ygbyy,... (omit the final ggggg)")
	cheat      = flag.Bool("cheat", false, "Use the actual game dicts instead of the open source")
)

// loadFile returns the contents of a file split on newlines, sorted, and uniqued
func loadFile(file string) []string {
	raw, _ := ioutil.ReadFile(file)
	return strings.Split(string(raw), "\n")
}

// loadDict returns the mystery and guessable word lists
func loadDicts() ([]string, []string) {
	if *cheat {
		return sortUnique(loadFile("mystery.txt")), sortUnique(loadFile("guessable.txt"))
	}

	// We have to explicitly make a copy. Otherwise one will be a reference
	// to the other and we will get data corruption when we try to manipulate
	// the dictionaries separately.
	mysteries := sortUnique(loadFile("../spellable/dict.huge"))
	guessables := make([]string, len(mysteries))
	copy(guessables, mysteries)

	return mysteries, guessables
}

// unpackMasks returns the command line masks representation in slices
func unpackMasks(s string) ([]string, error) {
	masks := strings.Split(s, ",")

	if len(masks) <= 1 {
		return nil, fmt.Errorf("Too few masks %v %d", masks, len(masks))
	}

	l := len(masks[0])
	for _, m := range masks {
		if len(m) != l {
			return nil, fmt.Errorf("Masks must all be of the same length %v %s", masks, m)
		}

		for _, val := range m {
			switch val {
			case 'g':
			case 'y':
			case 'b':
			default:
				return nil, fmt.Errorf("Masks must contain only g, y, or b %v %s %c", masks, m, val)
			}
		}
	}

	return sortUnique(masks), nil
}

// filterByLen returns all words of a given len from a given list of words
func filterByLen(words []string, l int) []string {
	matches := []string{}

	for _, word := range words {
		if len(word) == l {
			matches = append(matches, word)
		}
	}

	return matches
}

// matchSingleWord returns true if word matches candidate based on the given mask
func matchSingleWord(word, mask, candidate string) bool {
	if len(word) != len(mask) || len(word) != len(candidate) {
		fmt.Println("Internal consistency error!", word, mask, candidate)
		return false
	}

	// We will mask out some of the letters in word. Store that in w.
	w := ""

	// Evaluate 'b' and 'g' masks
	for i, m := range mask {
		switch m {
		case 'g':
			if word[i] != candidate[i] {
				return false
			}
			// This letter has been "spoken for", mark it as such
			w += "_"
			continue
		case 'b':
			if strings.ContainsRune(word, rune(candidate[i])) {
				return false
			}
		case '.':
		}

		w += string(word[i])
	}

	// Evaluate 'y' masks
	for i, m := range mask {
		if m != 'y' {
			continue
		}

		// The candidate letter must be in word...
		if !strings.ContainsRune(w, rune(candidate[i])) {
			return false
		}

		// ...but it can't be *this* letter
		if w[i] == candidate[i] {
			return false
		}

		// Mark the letter as having been "spoken for"
		w = strings.Replace(w, string(candidate[i]), "_", 1)
	}

	return true
}

func getLetters(word, mask, match string, masklet byte) string {
	l := ""

	for i := range mask {
		if mask[i] == masklet {
			l += string(match[i])
		}
	}

	return l
}

func greenMask(masks []string) string {
	green := ""

	for i := 0; i < len(masks[0]); i++ {
		foundGreen := false
		for _, mask := range masks {
			if mask[i] == 'g' {
				foundGreen = true
				green += string('g')
				break
			}
		}
		if !foundGreen {
			green += string('.')
		}
	}

	return green
}

// chainsHelper returns whether there exists even one valid chain of words that satisfy the masks
func chainsHelper(word string, masks []string, matches [][]string, depth int, chain []string) []string {
	// Can we add another word to the chain?
	for _, match := range matches[depth] {
		if !matchSingleWord(word, masks[depth], match) {
			continue
		}

		depth++

		c := append(chain, match)

		if depth >= len(masks) {
			return c
		}

		c = chainsHelper(word, masks, matches, depth, c)
		if c == nil {
			// We could not continue the chain with this word. Try the next word.
			continue
		}

		return c
	}

	return nil
}

func chains(word string, masks []string, matches [][]string) []string {
	return chainsHelper(word, masks, matches, 0, nil)
}

// matchMasks returns whether any candidate words match the word/masks pair
func matchMasks(word string, masks, candidates []string) bool {
	matches := make([][]string, len(masks))

	// Rule out the cases where candidates do not match the masks
	for i, mask := range masks {
		for _, candidate := range candidates {
			if !matchSingleWord(word, mask, candidate) {
				continue
			}
			matches[i] = append(matches[i], candidate)
		}

		if len(matches[i]) == 0 {
			return false
		}
	}

	// We have ruled out cases where an individual mask has no possible matches.
	// Now determine whether the remaining words provide any chains or whether
	// they are all unrelated. If there are no chains then this word cannot be
	// a solution.
	chain := chains(word, masks, matches)
	fmt.Println(word, chain)

	return chain != nil
}

func applyMasks(mysteries, guessables []string, masks []string) []string {
	matches := []string{}

	// For each candidate, find the matches for each mask
	for _, mystery := range mysteries {
		if matchMasks(mystery, masks, guessables) {
			matches = append(matches, mystery)
		}
	}

	return matches
}

func sortUnique(s []string) []string {
	// Make a copy so we do not corrupt the backing array of s
	s2 := make([]string, len(s))
	copy(s2, s)

	sort.Strings(s2)

	last := s2[0]
	for i := 1; i < len(s2); {
		if s2[i] == last {
			// Delete this duplicate
			s2 = append(s2[:i], s2[i+1:]...)
			continue
		}
		last = s2[i]
		i++
	}

	return s2
}

func letterFrequency(matches []string) (map[byte]int, []map[byte]int) {
	positions := len(matches[0])
	lFreq := map[byte]int{}
	lbpFreq := make([]map[byte]int, positions)

	for i := range lbpFreq {
		lbpFreq[i] = map[byte]int{}
	}

	for _, match := range matches {
		for i := 0; i < positions; i++ {
			lbpFreq[i][match[i]]++
			lFreq[match[i]]++
		}
	}

	return lFreq, lbpFreq
}

func prettyPrintFreq(f map[byte]int) string {
	out := []string{}

	for key, val := range f {
		str := fmt.Sprintf("%c:%2d", key, val)
		out = append(out, str)
	}

	return fmt.Sprintf("  %s\n", strings.Join(sortUnique(out), " "))
}

func printStats(matches, masks []string) {
	fmt.Println()

	samples := 10
	if samples > len(matches) {
		samples = len(matches)
	}
	fmt.Printf("Found %d matches for masks %v, printing first %d...\n", len(matches), masks, samples)
	fmt.Println(matches[:samples])

	lFreq, lByPos := letterFrequency(matches)

	fmt.Println("Letter frequency by position:")
	for i, pos := range lByPos {
		fmt.Printf("  [%d] %s\n", i, prettyPrintFreq(pos))
	}

	fmt.Println("Letter frequency overall:")
	fmt.Printf(prettyPrintFreq(lFreq))

	maxWord, maxScore := scoreWords(matches, lFreq)
	fmt.Printf("\nSuggested guess: '%s' for a score of %d\n", maxWord, maxScore)
}

// scoreWord returns the sum of unique letter frequencies for a given word
func scoreWord(word string, freq map[byte]int) int {
	used := map[rune]bool{}
	score := 0

	for _, val := range word {
		if used[val] {
			continue
		}
		score += freq[byte(val)]
		used[val] = true
	}

	return score
}

func scoreWords(words []string, lFreq map[byte]int) (string, int) {
	maxScore := 0
	maxWord := ""

	for _, word := range words {
		score := scoreWord(word, lFreq)
		if score > maxScore {
			maxScore = score
			maxWord = word
		}
	}

	return maxWord, maxScore
}

func crack(m string) error {
	masks, err := unpackMasks(m)
	if err != nil {
		return err
	}

	mysteries, guessables := loadDicts()

	// Use only the words of appropriate length
	mysteries = filterByLen(mysteries, len(masks[0]))
	guessables = filterByLen(guessables, len(masks[0]))

	// Find which mystery words can be formed using words from the guessable words
	matches := applyMasks(mysteries, guessables, masks)

	printStats(matches, masks)

	return nil
}

func main() {
	fmt.Printf("Welcome to Cracker\n\n")

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	err := crack(*masks)
	if err != nil {
		fmt.Println(err)
	}
}
