package seedpoet_test

import (
	"fmt"
	"strings"

	"github.com/abkvme/seed-poet/v2/seedpoet"
)

// Generate a fresh 24-word poem from crypto/rand. The poem differs on every
// run; read word by word it is always a valid BIP39 mnemonic.
func ExampleGenerate() {
	p, err := seedpoet.Generate(24)
	if err != nil {
		panic(err)
	}
	// fmt.Println(p.Format()) prints the 8 rhyming lines.
	words := p.Words()
	fmt.Println(len(p.Lines), "lines,", len(words), "words")
	fmt.Println("valid BIP39:", seedpoet.Validate(words) == nil)
	// Output:
	// 8 lines, 24 words
	// valid BIP39: true
}

// A fixed seed always gives the same poem. Use this only for tests and for
// checking another implementation, never for a real wallet.
func ExampleGenerateDeterministic() {
	p, err := seedpoet.GenerateDeterministic([]byte("example seed, never use for a wallet"), 12)
	if err != nil {
		panic(err)
	}
	fmt.Println(p.Format())
	// Output:
	// teach fiscal start
	// train rubber cart
	// spike long ship
	// say write clip
}

// The 24-word form: 8 lines of verb, adjective and rhyming noun.
func ExampleGenerateDeterministic_twentyFourWords() {
	p, err := seedpoet.GenerateDeterministic([]byte("example seed, never use for a wallet"), 24)
	if err != nil {
		panic(err)
	}
	fmt.Println(p.Format())
	// Output:
	// catch wrong sign
	// shop trash wine
	// exhibit mean control
	// plate electric hole
	// swarm team van
	// bind trip man
	// exercise large tree
	// alter merry knee
}

// Validate checks length, wordlist membership and the BIP39 checksum.
func ExampleValidate() {
	good := strings.Fields("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
	fmt.Println(seedpoet.Validate(good) == nil)

	badChecksum := strings.Fields("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon")
	fmt.Println(seedpoet.Validate(badChecksum) == nil)
	// Output:
	// true
	// false
}

// Wordlist returns a copy of the official BIP39 English list, in order.
func ExampleWordlist() {
	list := seedpoet.Wordlist()
	fmt.Println(len(list), list[0], list[len(list)-1])
	// Output:
	// 2048 abandon zoo
}
