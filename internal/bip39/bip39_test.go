package bip39

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/abkvme/seed-poet/v2/internal/wordlist"
)

func TestWordlistMatchesOfficialFile(t *testing.T) {
	// data/english.txt is the official BIP39 list
	// (sha256 2f5eed53a4727b4bf8880d8f3f199efc90e58503646d9ff8eff3a2ed3b24dbda).
	f, err := os.Open("../../data/english.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var words []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		words = append(words, strings.TrimSpace(sc.Text()))
	}
	if len(words) != 2048 || len(wordlist.BIP39Words) != 2048 {
		t.Fatalf("want 2048 words, file %d, Go list %d", len(words), len(wordlist.BIP39Words))
	}
	for i := range words {
		if words[i] != wordlist.BIP39Words[i] {
			t.Fatalf("word %d: file %q, Go list %q", i, words[i], wordlist.BIP39Words[i])
		}
	}
}

func TestValidateVectors(t *testing.T) {
	// Vectors from the BIP39 reference test set (trezor/python-mnemonic vectors.json).
	valid := []string{
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		"legal winner thank year wave sausage worth useful legal winner thank yellow",
		"letter advice cage absurd amount doctor acoustic avoid letter advice cage above",
		"zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
		"zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo vote",
		"ozone drill grab fiber curtain grace pudding thank cruise elder eight picnic",
		"hamster diagram private dutch cause delay private meat slide toddler razor book happy fancy gospel tennis maple dilemma loan word shrug inflict delay length",
	}
	for _, m := range valid {
		if err := Validate(strings.Fields(m)); err != nil {
			t.Errorf("%q: %v", m, err)
		}
	}
	invalid := []string{
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
		"zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon notaword",
	}
	for _, m := range invalid {
		if err := Validate(strings.Fields(m)); err == nil {
			t.Errorf("%q: accepted, want error", m)
		}
	}
}
