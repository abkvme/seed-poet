// Package syllable estimates English syllable counts with a vowel-group heuristic.
// The JavaScript port in js/seedpoet.js mirrors it exactly.
package syllable

import (
	"strings"
)

var vowels = "aeiouy"

// Count returns the estimated number of syllables in word (at least 1 for a
// non-empty word, 0 for an empty one).
func Count(word string) int {
	word = strings.ToLower(word)
	if len(word) == 0 {
		return 0
	}

	count := 0
	prevVowel := false

	for i, r := range word {
		isVowel := strings.ContainsRune(vowels, r)

		if isVowel && !prevVowel {
			count++
		}

		// Handle 'y' - only vowel if not at start
		if r == 'y' && i == 0 {
			count--
			if count < 0 {
				count = 0
			}
		}

		prevVowel = isVowel
	}

	// Silent-e adjustment
	if len(word) > 2 && strings.HasSuffix(word, "e") {
		// Check if it's not -le (which adds a syllable)
		if !strings.HasSuffix(word, "le") {
			count--
		}
	}

	// -ed ending is usually silent unless preceded by t or d
	if len(word) > 2 && strings.HasSuffix(word, "ed") {
		lastConsonant := word[len(word)-3]
		if lastConsonant != 't' && lastConsonant != 'd' {
			count--
		}
	}

	// Minimum 1 syllable
	if count < 1 {
		count = 1
	}

	return count
}

// CountAll returns the syllable count of every word, keyed by the word.
func CountAll(words []string) map[string]int {
	counts := make(map[string]int)
	for _, word := range words {
		counts[word] = Count(word)
	}
	return counts
}
