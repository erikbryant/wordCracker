package main

// go fmt && golint && go test && go run cracker.go -cheat=true -masks=bbbyy,yybbb -cpuprofile cpu.prof && echo top | go tool pprof cpu.prof

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
	cheat      = flag.Bool("cheat", false, "Use the actual game dicts instead of the open source")
	colorbars  = flag.String("colorbars", "bbbyy,gybbb,gyybb,gbygb,ggbgg", "colorbars from previous games in the form of yyybb,ygbyy,... (omit the final ggggg)")
	guessed    = flag.String("guessed", "", "comma-separated list of guess/colorbar pairs e.g., foo/gbb,oof/bby,...")
)

// loadFile returns the contents of a file split on newlines, sorted, and uniqued
func loadFile(file string) []string {
	raw, _ := ioutil.ReadFile(file)
	return strings.Split(string(raw), "\n")
}

// loadDict returns the mystery and guessable word lists
func loadDicts(cheat bool, wordLen int) ([]string, []string) {
	if cheat {
		mysteries := loadFile("../dictionaries/wordleMystery.dict")
		mysteries = sortUnique(mysteries)
		mysteries = filterByLen(mysteries, wordLen)

		guessables := loadFile("../dictionaries/wordleGuessable.dict")
		guessables = sortUnique(guessables)
		guessables = filterByLen(guessables, wordLen)

		return mysteries, guessables
	}

	// Even though they are identical, make a copy. Otherwise one will be a
	// reference to the other and we will get data corruption if we ever try
	// to manipulate the dictionaries separately.
	mysteries := loadFile("../dictionaries/huge.dict")
	mysteries = sortUnique(mysteries)
	mysteries = filterByLen(mysteries, wordLen)

	guessables := make([]string, len(mysteries))
	copy(guessables, mysteries)

	return mysteries, guessables
}

// validMask
func validMask(mask string, length int) (bool, error) {
	if len(mask) != length {
		return false, fmt.Errorf("Masks must all be of the same length %s", mask)
	}

	for _, val := range mask {
		switch val {
		case 'g':
		case 'y':
		case 'b':
		default:
			return false, fmt.Errorf("Masks must contain only g, y, or b %s %c", mask, val)
		}
	}

	return true, nil
}

// unpackMasks returns the command line masks representation in slices
func unpackMasks(s string) ([]string, error) {
	masks := strings.Split(s, ",")

	if len(masks) <= 1 {
		return nil, fmt.Errorf("Too few masks %v %d", masks, len(masks))
	}

	for _, m := range masks {
		if ok, err := validMask(m, len(masks[0])); !ok {
			return nil, err
		}
	}

	return sortUnique(masks), nil
}

// unpackGuessed returns the words guessed and the resulting colorbar masks in slices
func unpackGuessed(guessedPairs string) ([]string, []string, error) {
	guessWords := []string{}
	guessMasks := []string{}

	for _, pair := range strings.Split(guessedPairs, ",") {
		s := strings.Split(pair, "/")
		if len(s) != 2 {
			return nil, nil, fmt.Errorf("too many/few slash-delimited values %s %v", pair, s)
		}
		if ok, err := validMask(s[1], len(s[0])); !ok {
			return nil, nil, err
		}
		guessWords = append(guessWords, s[0])
		guessMasks = append(guessMasks, s[1])
	}

	return guessWords, guessMasks, nil
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

// replace replaces the first instance of a with b in w
func replace(w []byte, a, b byte) {
	for i := range w {
		if w[i] == a {
			w[i] = b
			return
		}
	}
}

// contains returns true if b is in w
func contains(w []byte, b byte) bool {
	for _, val := range w {
		if val == b {
			return true
		}
	}
	return false
}

// matchSingleWord returns true if candidate is not ruled out based on word/mask
func matchSingleWord(word, mask, candidate string) bool {
	if len(word) != len(mask) || len(word) != len(candidate) {
		fmt.Println("Internal consistency error!", word, mask, candidate)
		return false
	}

	// We will mask out some of the letters in word. Store that in w.
	w := make([]byte, len(word))

	// Evaluate 'b' and 'g' masks
	for i, m := range mask {
		switch m {
		case 'g':
			if word[i] != candidate[i] {
				return false
			}
			// This letter has been "spoken for", mark it as such
			w[i] = '_'
			continue
		case 'b':
			if strings.ContainsRune(word, rune(candidate[i])) {
				return false
			}
		case '.':
		}

		w[i] = word[i]
	}

	// Evaluate 'y' masks
	for i, m := range mask {
		if m != 'y' {
			continue
		}

		// The candidate letter must be in word...
		if !contains(w, candidate[i]) {
			return false
		}

		// ...but it can't be *this* letter
		if w[i] == candidate[i] {
			return false
		}

		// Mark the letter as having been "spoken for"
		replace(w, candidate[i], '_')
	}

	return true
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

	// temp := []string{}
	// for _, match := range matches {
	// 	temp = append(temp, match[0])
	// }
	// fmt.Println(word, temp)

	return true
}

// applyMasks returns the set of matches for a given set of masks
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

// sortUnique sorts a list and removes any duplicates
func sortUnique(s []string) []string {
	if len(s) <= 0 {
		return []string{}
	}

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

// letterFrequency returns maps of the frequency of letters in the given matches
func letterFrequency(matches []string) ([]int, [][]int) {
	if len(matches) == 0 {
		return nil, nil
	}

	letterLen := 256 // This is too many, but it is fast
	positions := len(matches[0])
	lFreq := make([]int, letterLen)
	lbpFreq := make([][]int, positions)

	for i := range lbpFreq {
		lbpFreq[i] = make([]int, letterLen)
	}

	for _, match := range matches {
		for i := 0; i < positions; i++ {
			lbpFreq[i][match[i]]++
			lFreq[match[i]]++
		}
	}

	return lFreq, lbpFreq
}

// prettyPrintFreq returns a formatted string representation of the given map
func prettyPrintFreq(f []int) string {
	out := []string{}

	for key, val := range f {
		if val == 0 {
			continue
		}
		str := fmt.Sprintf("%c:%2d", key, val)
		out = append(out, str)
	}

	return fmt.Sprintf("  %s\n", strings.Join(sortUnique(out), " "))
}

// scoreWord returns the sum of unique letter frequencies for a given word
func scoreWord(word string, freq []int) int {
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

type score struct {
	score int
	word  string
}

// scoreWords returns the scores for each word and the words with the max score
func scoreWords(words []string, lFreq []int) ([]string, int, []score) {
	maxScore := 0
	maxWords := []string{}
	scores := make([]score, len(words))

	for i, word := range words {
		score := scoreWord(word, lFreq)

		scores[i].score = score
		scores[i].word = word

		if score > maxScore {
			maxScore = score
			maxWords = []string{word}
			continue
		}

		if score == maxScore {
			maxWords = append(maxWords, word)
		}
	}

	return maxWords, maxScore, scores
}

// printStats prints statistics abut the matches
func printStats(matches, masks []string, message string) {
	fmt.Println()
	fmt.Println("===================================================")

	if message != "" {
		fmt.Println(message)
		fmt.Println()
	}

	samples := 10
	if samples > len(matches) {
		samples = len(matches)
	}
	fmt.Printf("Found %d matches for masks %v, printing first few...\n", len(matches), masks)
	fmt.Println(matches[:samples])

	lFreq, lByPos := letterFrequency(matches)

	fmt.Println("Letter frequency by position:")
	for i, pos := range lByPos {
		fmt.Printf("  [%d] %s\n", i, prettyPrintFreq(pos))
	}

	fmt.Println("Letter frequency overall:")
	fmt.Printf(prettyPrintFreq(lFreq))

	maxWords, maxScore, _ := scoreWords(matches, lFreq)
	fmt.Printf("\nSuggested guess(es): %v for a score of %d\n", maxWords, maxScore)

	fmt.Println("===================================================")
	fmt.Println()
}

// crack runs the main loop
func crack(mysteries, guessables, masks []string) error {
	// Find which mystery words can be formed using words from the guessable words
	matches := applyMasks(mysteries, guessables, masks)

	printStats(matches, masks, "")

	return nil
}

// makeMask returns the byg mask for the given guess and given mystery word
func makeMask(word, guess string) string {
	m := make([]byte, len(word))
	w := make([]byte, len(word))

	for i, val := range word {
		w[i] = byte(val)
	}

	// g
	for i := range word {
		if word[i] != guess[i] {
			continue
		}
		m[i] = 'g'
		w[i] = '_'
	}

	// y and b
	for i := range word {
		if m[i] != 0 {
			continue
		}

		if contains(w, guess[i]) {
			m[i] = 'y'
			replace(w, guess[i], '_')
			continue
		}

		m[i] = 'b'
	}

	mask := ""
	for _, val := range m {
		mask += string(val)
	}

	return mask
}

// findMaxScore returns the highest scoring word that has not already been guessed
func findMaxScore(scores []score, guesses string) score {
	max := score{-1, ""}

	for _, s := range scores {
		if s.score > max.score && !strings.Contains(guesses, s.word) {
			max.score = s.score
			max.word = s.word
		}
	}

	return max
}

func suggestGuess(matches []string, guesses string) string {
	lFreq, _ := letterFrequency(matches)
	_, _, scores := scoreWords(matches, lFreq)
	max := findMaxScore(scores, guesses)

	return max.word
}

func pruneGuessables(guessables []string, word, mask string) []string {
	pruned := []string{}

	//
	// Yellow
	//
	yellowLetters := []byte{}
	yellowLetterPos := [][]int{}
	for i := range mask {
		if mask[i] != 'y' {
			continue
		}

		yellowLetters = append(yellowLetters, word[i])
		yellowLetterPos = append(yellowLetterPos, []int{})
		index := len(yellowLetters) - 1

		for j := range mask {
			if j == i {
				// Skip our own letter
				continue
			}
			if mask[j] == 'g' {
				continue
			}
			// This is y or b and *could* be where our y-letter goes
			yellowLetterPos[index] = append(yellowLetterPos[index], j)
		}
	}

	//
	// Black
	//
	blackLetters := ""
	for i := range mask {
		if mask[i] != 'b' {
			continue
		}

		solvable := true
		for j := range mask {
			if mask[j] == 'y' && word[j] == word[i] {
				// We do not have enough information to solve. Abort.
				solvable = false
				break
			}
		}
		if !solvable {
			continue
		}

		// This is a b letter and there are no y letters that are the same as this
		// so this letter can't be anywhere in the word
		blackLetters += string(word[i])
	}

	for _, guess := range guessables {
		if !greenCompliant(word, mask, guess) {
			continue
		}

		if !yellowCompliant(yellowLetters, yellowLetterPos, guess) {
			continue
		}

		if !blackCompliant(blackLetters, mask, guess) {
			continue
		}

		pruned = append(pruned, guess)
	}

	return pruned
}

func greenCompliant(word, mask, guess string) bool {
	for i := range mask {
		if mask[i] == 'g' && guess[i] != word[i] {
			return false
		}
	}

	return true
}

func yellowCompliant(letters []byte, letterPos [][]int, guess string) bool {
	if len(letters) == 0 {
		return true
	}

	for i, val := range letters {
		// We need to find at least one instance of this letter
		for _, pos := range letterPos[i] {
			if guess[pos] == val {
				return true
			}
		}
	}

	return false
}

func blackCompliant(letters string, mask, guess string) bool {
	for i := range mask {
		if mask[i] == 'g' {
			continue
		}
		if strings.ContainsAny(string(guess[i]), letters) {
			return false
		}
	}

	return true
}

func playAllWords(wordLen int) {
	mysteries, guessables := loadDicts(false, wordLen)

	totalGuesses := 0
	totalWords := 0

	for _, mystery := range mysteries {
		masks := []string{}
		guesses := ""
		guessableWords := guessables
		totalWords++

		for i := 1; ; i++ {
			guess := suggestGuess(guessableWords, guesses)
			guesses += "." + guess
			totalGuesses++

			mask := makeMask(mystery, guess)
			masks = append(masks, mask)

			if guess == mystery {
				break
			}

			guessableWords = pruneGuessables(guessableWords, guess, mask)
		}
	}

	fmt.Println("Total words:", totalWords)
	fmt.Println("Average guesses:", float64(totalGuesses)/float64(totalWords))
}

func solveOne(mysteries, guessables, masks, guessWords, guessMasks []string) {
	// Find which mystery words can be formed using words from the guessable words
	matches := applyMasks(mysteries, guessables, masks)
	printStats(matches, masks, "Analysis of initial masks")

	guesses := ""
	for i := range guessWords {
		masks = append(masks, guessMasks[i])
		guesses += "." + guessWords[i]

		matches = pruneGuessables(matches, guessWords[i], guessMasks[i])
		msg := fmt.Sprintf("After applying %s/%s", guessWords[i], guessMasks[i])
		printStats(matches, masks, msg)
		fmt.Println(matches)
	}

	guess := suggestGuess(matches, guesses)
	fmt.Println("Suggested guess:", guess)
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

	// playAllWords(5)
	// return

	masks, err := unpackMasks(*colorbars)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Use only the words of appropriate length
	wordLen := len(masks[0])
	mysteries, guessables := loadDicts(*cheat, wordLen)

	if *guessed == "" {
		err = crack(mysteries, guessables, masks)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		guessWords, guessMasks, err := unpackGuessed(*guessed)
		if err != nil {
			fmt.Println(err)
			return
		}
		solveOne(mysteries, guessables, masks, guessWords, guessMasks)
	}
}
