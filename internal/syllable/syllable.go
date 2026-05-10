package syllable

import (
	"strings"
)

var vowels = "aeiouy"

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

	// -es ending is usually silent
	if len(word) > 3 && strings.HasSuffix(word, "es") {
		if !strings.ContainsRune(vowels, rune(word[len(word)-3])) {
			// consonant + es, don't count extra
		}
	}

	// -ed ending is usually silent unless preceded by t or d
	if len(word) > 2 && strings.HasSuffix(word, "ed") {
		lastConsonant := word[len(word)-3]
		if lastConsonant != 't' && lastConsonant != 'd' {
			count--
		}
	}

	// -ion is one syllable, not two
	if strings.HasSuffix(word, "tion") || strings.HasSuffix(word, "sion") {
		// Already counted correctly by vowel grouping
	}

	// Minimum 1 syllable
	if count < 1 {
		count = 1
	}

	return count
}

// CountAll returns syllable counts for a list of words
func CountAll(words []string) map[string]int {
	counts := make(map[string]int)
	for _, word := range words {
		counts[word] = Count(word)
	}
	return counts
}
