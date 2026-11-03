// Package bip39 validates mnemonics against the BIP39 English wordlist,
// including the checksum carried in the last word.
package bip39

import (
	"crypto/sha256"
	"fmt"

	"github.com/abkvme/seed-poet/v2/internal/wordlist"
)

var index = func() map[string]int {
	m := make(map[string]int, len(wordlist.BIP39Words))
	for i, w := range wordlist.BIP39Words {
		m[w] = i
	}
	return m
}()

// Index returns the position of word in the BIP39 English wordlist.
func Index(word string) (int, bool) {
	i, ok := index[word]
	return i, ok
}

// Validate reports whether words form a valid BIP39 mnemonic: a supported
// length (12, 15, 18, 21 or 24), every word in the list, and a correct checksum.
func Validate(words []string) error {
	n := len(words)
	if n < 12 || n > 24 || n%3 != 0 {
		return fmt.Errorf("bip39: unsupported word count %d", n)
	}
	totalBits := n * 11
	csBits := totalBits / 33
	entBits := totalBits - csBits

	bits := make([]byte, 0, totalBits)
	for _, w := range words {
		i, ok := index[w]
		if !ok {
			return fmt.Errorf("bip39: %q is not in the wordlist", w)
		}
		for b := 10; b >= 0; b-- {
			bits = append(bits, byte(i>>b)&1)
		}
	}

	entropy := make([]byte, entBits/8)
	for i := 0; i < entBits; i++ {
		entropy[i/8] |= bits[i] << (7 - i%8)
	}
	sum := sha256.Sum256(entropy)
	for i := 0; i < csBits; i++ {
		if (sum[i/8]>>(7-i%8))&1 != bits[entBits+i] {
			return fmt.Errorf("bip39: checksum mismatch")
		}
	}
	return nil
}
