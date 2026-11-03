// Package wordlist exposes the BIP39 list and the Seed Poet word categories
// loaded from data/seedpoet.json.
package wordlist

import "github.com/abkvme/seed-poet/data"

var (
	// BIP39Words contains all 2048 BIP39 English mnemonic words, in order.
	BIP39Words = data.BIP39
	// Verbs are action words for the first position of a line.
	Verbs = data.Verbs
	// Adjectives are describing words for the middle position.
	Adjectives = data.Adjectives
	// Nouns are things for the end position (the rhyme).
	Nouns = data.Nouns
)
