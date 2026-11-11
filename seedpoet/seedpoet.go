// Package seedpoet generates BIP39 seed phrases written as rhyming couplets.
//
// Every poem, read word by word, is a valid BIP39 English mnemonic with a
// correct checksum, so any BIP39 wallet accepts it. Randomness comes from
// crypto/rand. The JavaScript library in js/seedpoet.js implements the same
// algorithm and produces the same poem from the same random bytes.
//
//	p, err := seedpoet.Generate(24)
//	if err != nil { ... }
//	fmt.Println(p.Format())  // 8 rhyming lines
//	words := p.Words()       // the 24-word mnemonic
package seedpoet

import (
	"sync"

	"github.com/abkvme/seed-poet/v2/internal/bip39"
	"github.com/abkvme/seed-poet/v2/internal/generator"
	"github.com/abkvme/seed-poet/v2/internal/rhyme"
	"github.com/abkvme/seed-poet/v2/internal/rng"
	"github.com/abkvme/seed-poet/v2/internal/wordlist"
)

// Version of Seed Poet (Go library, CLI, website). The JavaScript library
// (js/package.json) carries the same number; a test keeps them equal.
const Version = "2.1.3"

// Poem is a generated poem: Lines of three words (verb, adjective, rhyming noun).
type Poem = generator.Poem

var (
	once sync.Once
	gen  *generator.Generator
)

func defaultGenerator() *generator.Generator {
	once.Do(func() {
		gen = generator.New(wordlist.Verbs, wordlist.Adjectives, wordlist.Nouns, rhyme.BuildPhoneticIndex())
	})
	return gen
}

// Generate returns a new poem of 12 or 24 words (other values give 12).
// It is safe for concurrent use.
func Generate(words int) (*Poem, error) {
	return defaultGenerator().Generate(words)
}

// GenerateDeterministic returns the poem for a fixed seed. It exists for tests
// and for checking that another implementation matches this one. Never use it
// to create a real wallet seed: anyone with the seed gets the same poem.
func GenerateDeterministic(seed []byte, words int) (*Poem, error) {
	return defaultGenerator().GenerateWith(rng.Deterministic(seed), words)
}

// Validate reports whether words form a valid BIP39 English mnemonic
// (supported length, known words, correct checksum).
func Validate(words []string) error {
	return bip39.Validate(words)
}

// Wordlist returns the official BIP39 English wordlist (2048 words, in order).
func Wordlist() []string {
	return append([]string(nil), wordlist.BIP39Words...)
}
