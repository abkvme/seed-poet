package main

import (
	"fmt"
	"log"
	"os"

	"github.com/abkvme/seed-poet/internal/generator"
	"github.com/abkvme/seed-poet/internal/rhyme"
	"github.com/abkvme/seed-poet/internal/server"
	"github.com/abkvme/seed-poet/internal/wordlist"
)

const (
	appName    = "Seed Poet"
	appVersion = "1.2.0"
)

func main() {
	// Build phonetic rhyme index
	rhymeIndex := rhyme.BuildPhoneticIndex()

	// Create generator with VERB + ADJECTIVE + NOUN structure
	gen := generator.New(
		wordlist.Verbs,
		wordlist.Adjectives,
		wordlist.Nouns,
		rhymeIndex,
	)

	args := os.Args[1:]

	if len(args) == 0 {
		runServer(gen, ":8080")
		return
	}

	switch args[0] {
	case "12":
		runCLI(gen, 12)
	case "24":
		runCLI(gen, 24)
	case "serve":
		port := ":8080"
		if len(args) > 1 {
			port = args[1]
		}
		runServer(gen, port)
	default:
		printUsage()
	}
}

func runCLI(gen *generator.Generator, wordCount int) {
	fmt.Printf("%s v%s\n\n", appName, appVersion)
	poem, err := gen.Generate(wordCount)
	if err != nil {
		log.Fatalf("Generate: %v", err)
	}
	fmt.Println(poem.Format())
}

func runServer(gen *generator.Generator, addr string) {
	srv := server.New(gen)
	log.Printf("%s v%s starting on http://localhost%s", appName, appVersion, addr)
	if err := srv.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func printUsage() {
	fmt.Printf("%s v%s\n", appName, appVersion)
	fmt.Println("Generate memorable poems using BIP39 mnemonic words")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  seedpoet              Start web server on :8080")
	fmt.Println("  seedpoet 12           Generate 12-word poem")
	fmt.Println("  seedpoet 24           Generate 24-word poem")
	fmt.Println("  seedpoet serve        Start web server on :8080")
	fmt.Println("  seedpoet serve :3000  Start on custom port")
}
