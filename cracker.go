package main

// go fmt ./... && go vet ./... && go test ./... && go run cracker.go -colorbars=bbbyy,yybbb -cpuprofile cpu.prof && echo top | go tool pprof cpu.prof

import (
	"flag"
	"fmt"
	"github.com/erikbryant/dictionaries"
	"log"
	"os"
	"runtime/pprof"
	"strings"
)

var (
	cpuprofile  = flag.String("cpuprofile", "", "write cpu profile to file")
	colorbars   = flag.String("colorbars", "ggggg", "colorbars from previous games in the form of yyybb,ygbyy,... (omit the final ggggg)")
	guessed     = flag.String("guessed", "", "comma-separated list of guess/colorbar pairs e.g., foo/gbb,oof/bby,...")
	mysteryWord = flag.String("mystery", "", "the mystery word (if you know it), useful for error checking masks")
)

// loadDicts returns the mystery and guessable word lists
func loadDicts(wordLen int) ([]string, []string) {
	// Even though they are identical, make a copy. Otherwise one will be a
	// reference to the other and we will get data corruption if we ever try
	// to manipulate the dictionaries separately.
	mysteries := dictionaries.LoadFile("../dictionaries/merged.dict")
	mysteries = dictionaries.FilterByLen(mysteries, wordLen)
	mysteries = dictionaries.SortUnique(mysteries)

	guessables := make([]string, len(mysteries))
	copy(guessables, mysteries)

	return mysteries, guessables
}

// validMask returns true if the mask appears to be valid
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

	for _, m := range masks {
		if ok, err := validMask(m, len(masks[0])); !ok {
			return nil, err
		}
	}

	return dictionaries.SortUnique(masks), nil
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

	// Evaluate 'g' masks
	for i, m := range mask {
		switch m {
		case 'g':
			if word[i] != candidate[i] {
				return false
			}
			// This letter has been "spoken for", mark it as such
			w[i] = '_'
			continue
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

	// Evaluate 'b' masks
	for i, m := range mask {
		switch m {
		case 'b':
			// Only look at letters that are not already spoken for
			if strings.ContainsRune(string(w), rune(candidate[i])) {
				return false
			}
		}
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

	return true
}

// applyMasks returns the set of matches for a given set of masks
func applyMasks(mysteries, guessables, masks []string) []string {
	matches := []string{}

	// For each candidate, find the matches for each mask
	for _, mystery := range mysteries {
		if matchMasks(mystery, masks, guessables) {
			matches = append(matches, mystery)
		}
	}

	return matches
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

	lFreq, lByPos := dictionaries.LetterFrequency(matches)

	fmt.Println("Letter frequency by position:")
	for i, pos := range lByPos {
		fmt.Printf("  [%d] %s\n", i, dictionaries.PrettyPrintFreq(pos))
	}

	fmt.Println("Letter frequency overall:")
	fmt.Printf(dictionaries.PrettyPrintFreq(lFreq))

	maxWords, maxScore, _ := scoreWords(matches, lFreq)
	fmt.Printf("\nSuggested guess(es): %v for a score of %d\n", maxWords, maxScore)

	fmt.Println("===================================================")
	fmt.Println()
}

// crack eliminates all wods that do not match the masks
func crack(mysteries, guessables, masks []string, mystery string) error {
	// Find which mystery words can be formed using words from the guessable words
	matches := applyMasks(mysteries, guessables, masks)
	if mystery != "" && !dictionaries.ContainsWord(matches, mystery) {
		return fmt.Errorf("mystery word is not in matches %v %s", matches, mystery)
	}

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

func suggestGuessLetterFreq(matches []string, guesses string) string {
	lFreq, _ := dictionaries.LetterFrequency(matches)
	_, _, scores := scoreWords(matches, lFreq)
	max := findMaxScore(scores, guesses)

	return max.word
}

func pruneGuessables(guessables []string, word, mask string) []string {
	pruned := []string{}

	for _, guess := range guessables {
		guessMask := makeMask(guess, word)
		if guessMask == mask {
			pruned = append(pruned, guess)
		}
	}

	return pruned
}

func playAllWords(wordLen int) {
	mysteries, guessables := loadDicts(wordLen)

	totalGuesses := 0
	totalWords := 0

	for _, mystery := range mysteries {
		masks := []string{}
		guesses := ""
		guessableWords := guessables
		totalWords++

		for i := 1; ; i++ {
			guess := suggestGuessLetterFreq(guessableWords, guesses)
			guesses += "." + guess
			totalGuesses++

			mask := makeMask(mystery, guess)
			masks = append(masks, mask)

			if guess == mystery {
				fmt.Printf("Mystery: %s  Guesses: %2d  Total Words: %5d  Average guesses: %4.2f\n", mystery, i, totalWords, float64(totalGuesses)/float64(totalWords))
				break
			}

			guessableWords = pruneGuessables(guessableWords, guess, mask)
		}
	}

	fmt.Printf("\nTotal Words: %5d  Average guesses: %4.2f\n", totalWords, float64(totalGuesses)/float64(totalWords))
}

func solveOne(mysteries, guessables, masks, guessWords, guessMasks []string, mystery string) error {
	// Find which mystery words can be formed using words from the guessable words
	matches := applyMasks(mysteries, guessables, masks)
	if mystery != "" && !dictionaries.ContainsWord(matches, mystery) {
		return fmt.Errorf("mystery word has been excluded from matches %v %s", matches, mystery)
	}
	printStats(matches, masks, "Analysis of initial masks")

	guesses := ""
	for i := range guessWords {
		masks = append(masks, guessMasks[i])
		guesses += "." + guessWords[i]

		matches = pruneGuessables(matches, guessWords[i], guessMasks[i])
		if mystery != "" && !dictionaries.ContainsWord(matches, mystery) {
			return fmt.Errorf("mystery word: '%s' has been excluded from matches after guessing: '%s'. %v", mystery, guessWords[i], matches)
		}
		msg := fmt.Sprintf("After applying %s/%s", guessWords[i], guessMasks[i])
		printStats(matches, masks, msg)
		fmt.Println(matches)
	}

	fmt.Println("===================================================")
	guess := suggestGuessLetterFreq(matches, guesses)
	fmt.Println("Suggested guess:", guess)
	fmt.Println("===================================================")
	fmt.Println()

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

	// playAllWords(5)
	// return

	masks, err := unpackMasks(*colorbars)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Use only the words of appropriate length
	wordLen := len(masks[0])
	mysteries, guessables := loadDicts(wordLen)

	// If there are no guesses, just find the set of matches
	if *guessed == "" {
		err = crack(mysteries, guessables, masks, *mysteryWord)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	// There is a guess. Unpack those and start solving.
	guessWords, guessMasks, err := unpackGuessed(*guessed)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = solveOne(mysteries, guessables, masks, guessWords, guessMasks, *mysteryWord)
	if err != nil {
		fmt.Println()
		fmt.Println("******** ERROR ********")
		fmt.Println()
		fmt.Println(err)
		return
	}
}
