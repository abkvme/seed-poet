package generator

import (
	"errors"
	"sort"
	"strings"

	"github.com/abkvme/seed-poet/v2/internal/bip39"
	"github.com/abkvme/seed-poet/v2/internal/rhyme"
	"github.com/abkvme/seed-poet/v2/internal/rng"
	"github.com/abkvme/seed-poet/v2/internal/syllable"
)

// wordPair holds a pre-computed verb+adjective combination
type wordPair struct {
	verb string
	adj  string
}

type Generator struct {
	verbs      []string // Position 1
	adjectives []string // Position 2
	nouns      []string // Position 3 (rhyming)
	rhymeIndex *rhyme.Index
	syllables  map[string]int

	// Pre-computed lookup tables for O(1) syllable matching
	verbsBySyllable map[int][]string   // syllables -> []verbs
	adjsBySyllable  map[int][]string   // syllables -> []adjectives
	comboBySyllable map[int][]wordPair // total_syllables -> []pairs
}

type Poem struct {
	Lines     [][]string
	WordCount int
}

func New(verbs, adjectives, nouns []string, rhymeIndex *rhyme.Index) *Generator {
	// Build syllable map for all words
	allWords := make([]string, 0, len(verbs)+len(adjectives)+len(nouns))
	allWords = append(allWords, verbs...)
	allWords = append(allWords, adjectives...)
	allWords = append(allWords, nouns...)
	syllables := syllable.CountAll(allWords)

	// Pre-compute verbs grouped by syllable count
	verbsBySyllable := make(map[int][]string)
	for _, v := range verbs {
		s := syllables[v]
		verbsBySyllable[s] = append(verbsBySyllable[s], v)
	}

	// Pre-compute adjectives grouped by syllable count
	adjsBySyllable := make(map[int][]string)
	for _, a := range adjectives {
		s := syllables[a]
		adjsBySyllable[s] = append(adjsBySyllable[s], a)
	}

	// Pre-compute all verb+adjective combinations by total syllable count.
	// Syllable counts are walked in ascending order (never Go map order) so the
	// pair lists are identical in Go and JavaScript.
	comboBySyllable := make(map[int][]wordPair)
	for _, vSyll := range sortedKeys(verbsBySyllable) {
		vList := verbsBySyllable[vSyll]
		for _, aSyll := range sortedKeys(adjsBySyllable) {
			aList := adjsBySyllable[aSyll]
			total := vSyll + aSyll
			for _, v := range vList {
				for _, a := range aList {
					if v != a { // Avoid same word as verb and adj
						comboBySyllable[total] = append(comboBySyllable[total], wordPair{v, a})
					}
				}
			}
		}
	}

	return &Generator{
		verbs:           verbs,
		adjectives:      adjectives,
		nouns:           nouns,
		rhymeIndex:      rhymeIndex,
		syllables:       syllables,
		verbsBySyllable: verbsBySyllable,
		adjsBySyllable:  adjsBySyllable,
		comboBySyllable: comboBySyllable,
	}
}

// maxAttempts bounds the checksum search. A random poem has a valid BIP39
// checksum with probability 1/16 (12 words) or 1/256 (24 words), so the
// chance of exhausting this many attempts is negligible (below 2^-50).
const maxAttempts = 10000

// ErrNoValidPoem is returned if no poem with a valid checksum was found.
var ErrNoValidPoem = errors.New("generator: no valid BIP39 poem found")

// Generate returns a poem whose words, read in order, form a valid BIP39
// mnemonic (checksum included). Unsupported word counts fall back to 12.
// All randomness comes from crypto/rand.
func (g *Generator) Generate(wordCount int) (*Poem, error) {
	r := rng.Crypto()
	defer r.Wipe()
	return g.GenerateWith(r, wordCount)
}

// GenerateWith draws poems from r until one has a valid BIP39 checksum, which
// keeps the choice uniform among the valid poems the generator can produce.
// With rng.Deterministic it is reproducible and matches the JavaScript library
// poem for poem (see testdata/vectors.json).
func (g *Generator) GenerateWith(r *rng.RNG, wordCount int) (*Poem, error) {
	if wordCount != 12 && wordCount != 24 {
		wordCount = 12
	}
	for i := 0; i < maxAttempts; i++ {
		p := g.generateOnce(r, wordCount)
		if bip39.Validate(p.Words()) == nil {
			return p, nil
		}
	}
	return nil, ErrNoValidPoem
}

func (g *Generator) generateOnce(r *rng.RNG, wordCount int) *Poem {
	var linesPerPoem, wordsPerLine, coupletCount int

	switch wordCount {
	case 12:
		linesPerPoem = 4
		wordsPerLine = 3
		coupletCount = 2
	case 24:
		linesPerPoem = 8
		wordsPerLine = 3
		coupletCount = 4
	default:
		linesPerPoem = 4
		wordsPerLine = 3
		coupletCount = 2
		wordCount = 12
	}

	used := make(map[string]bool)
	lines := make([][]string, 0, linesPerPoem)

	// Copy: the index is shared between concurrent requests and must not be shuffled in place.
	validKeys := append([]string(nil), g.rhymeIndex.GetValidKeys()...)
	if len(validKeys) == 0 {
		return &Poem{Lines: [][]string{}, WordCount: 0}
	}

	for len(validKeys) < coupletCount {
		validKeys = append(validKeys, validKeys...)
	}

	r.Shuffle(len(validKeys), func(i, j int) { validKeys[i], validKeys[j] = validKeys[j], validKeys[i] })
	selectedKeys := validKeys[:coupletCount]

	for _, key := range selectedKeys {
		group := g.rhymeIndex.GetGroup(key)
		if len(group) < 2 {
			continue
		}

		// Pick 2 rhyming nouns with similar syllable counts
		rhymeWords := g.pickTwoRhymingWithSimilarSyllables(r, group, used)
		if len(rhymeWords) < 2 {
			rhymeWords = g.pickTwoFromGroup(r, group)
		}

		// Mark both rhyme words as used BEFORE building lines
		// This prevents them from being picked as verbs/adjectives
		used[rhymeWords[0]] = true
		used[rhymeWords[1]] = true

		// Build line 1 and get its syllable count
		line1, syllCount1 := g.buildLineWithSyllables(r, rhymeWords[0], wordsPerLine, used)
		lines = append(lines, line1)

		// Build line 2 matching line 1's syllable count
		line2 := g.buildLineMatchingSyllables(r, rhymeWords[1], wordsPerLine, used, syllCount1)
		lines = append(lines, line2)
	}

	return &Poem{
		Lines:     lines,
		WordCount: wordCount,
	}
}

// buildLineWithSyllables builds a line: VERB + ADJECTIVE + NOUN
func (g *Generator) buildLineWithSyllables(r *rng.RNG, noun string, wordCount int, used map[string]bool) ([]string, int) {
	words := make([]string, wordCount)

	// Position 3 (last): noun for rhyming
	words[2] = noun
	used[noun] = true
	totalSyllables := g.syllables[noun]

	// Position 1 (first): verb
	verb := g.pickUnusedFrom(r, g.verbs, used)
	words[0] = verb
	used[verb] = true
	totalSyllables += g.syllables[verb]

	// Position 2 (middle): adjective
	adj := g.pickUnusedFrom(r, g.adjectives, used)
	words[1] = adj
	used[adj] = true
	totalSyllables += g.syllables[adj]

	return words, totalSyllables
}

// buildLineMatchingSyllables builds a line matching target syllable count
func (g *Generator) buildLineMatchingSyllables(r *rng.RNG, noun string, wordCount int, used map[string]bool, targetSyllables int) []string {
	words := make([]string, wordCount)

	// Position 3: noun
	words[2] = noun
	used[noun] = true
	nounSyllables := g.syllables[noun]

	// Calculate remaining syllables for verb + adjective
	remainingSyllables := targetSyllables - nounSyllables

	// Find verb + adjective combo that matches target
	verb, adj := g.pickVerbAndAdjForSyllables(r, remainingSyllables, used)
	words[0] = verb
	words[1] = adj
	used[verb] = true
	used[adj] = true

	return words
}

func (g *Generator) pickVerbAndAdjForSyllables(r *rng.RNG, targetTotal int, used map[string]bool) (string, string) {
	// Try exact match first, then close matches
	for _, target := range []int{targetTotal, targetTotal - 1, targetTotal + 1} {
		if pairs := g.comboBySyllable[target]; len(pairs) > 0 {
			// Try random pairs until we find unused ones (usually first try)
			for attempts := 0; attempts < 20; attempts++ {
				p := pairs[r.IntN(len(pairs))]
				if !used[p.verb] && !used[p.adj] {
					return p.verb, p.adj
				}
			}
		}
	}

	// Fallback: pick any unused verb and adjective
	return g.pickUnusedFrom(r, g.verbs, used), g.pickUnusedFrom(r, g.adjectives, used)
}

func (g *Generator) pickTwoRhymingWithSimilarSyllables(r *rng.RNG, group []string, used map[string]bool) []string {
	var available []string
	for _, word := range group {
		if !used[word] {
			available = append(available, word)
		}
	}

	if len(available) < 2 {
		return available
	}

	// Shuffle first to ensure randomness
	r.Shuffle(len(available), func(i, j int) { available[i], available[j] = available[j], available[i] })

	// Group by syllable count
	bySyllables := make(map[int][]string)
	for _, word := range available {
		syll := g.syllables[word]
		bySyllables[syll] = append(bySyllables[syll], word)
	}

	// Collect syllable counts that have 2+ words, in ascending order
	var validCounts []int
	for _, count := range sortedKeys(bySyllables) {
		if len(bySyllables[count]) >= 2 {
			validCounts = append(validCounts, count)
		}
	}

	// Pick a random valid syllable count
	if len(validCounts) > 0 {
		count := validCounts[r.IntN(len(validCounts))]
		words := bySyllables[count]
		r.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
		return words[:2]
	}

	// Fallback: any two from shuffled available
	return available[:2]
}

func (g *Generator) pickUnusedFrom(r *rng.RNG, wordList []string, used map[string]bool) string {
	n := len(wordList)
	// Try random index selection with retry (fast when few words are used)
	for attempts := 0; attempts < 50; attempts++ {
		word := wordList[r.IntN(n)]
		if !used[word] {
			return word
		}
	}
	// Fallback: linear scan (only if many words used)
	for _, word := range wordList {
		if !used[word] {
			return word
		}
	}
	return wordList[r.IntN(n)]
}

func (g *Generator) pickTwoFromGroup(r *rng.RNG, group []string) []string {
	if len(group) < 2 {
		return group
	}

	shuffled := make([]string, len(group))
	copy(shuffled, group)
	r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	return []string{shuffled[0], shuffled[1]}
}

func sortedKeys[V any](m map[int]V) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (p *Poem) Format() string {
	var lines []string
	for _, line := range p.Lines {
		lines = append(lines, strings.Join(line, " "))
	}
	return strings.Join(lines, "\n")
}

func (p *Poem) Words() []string {
	var words []string
	for _, line := range p.Lines {
		words = append(words, line...)
	}
	return words
}
