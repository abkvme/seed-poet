// Command seedpoet prints BIP39 seed phrases as rhyming poems, or serves the
// Seed Poet website.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/abkvme/seed-poet/v2/internal/server"
	"github.com/abkvme/seed-poet/v2/seedpoet"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		return
	}
	switch args[0] {
	case "12", "24":
		n := 12
		if args[0] == "24" {
			n = 24
		}
		p, err := seedpoet.Generate(n)
		if err != nil {
			log.Fatalf("generate: %v", err)
		}
		fmt.Printf("Seed Poet v%s\n\n%s\n", seedpoet.Version, p.Format())
	case "serve":
		addr := "127.0.0.1:8080"
		if len(args) > 1 {
			addr = args[1]
		}
		s, err := server.New()
		if err != nil {
			log.Fatalf("server: %v", err)
		}
		log.Fatal(s.Run(addr))
	case "version", "-v", "--version":
		fmt.Println(seedpoet.Version)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Printf(`Seed Poet v%s
BIP39 seed phrases written as rhyming poems.

Usage:
  seedpoet 12                    Print a 12-word poem
  seedpoet 24                    Print a 24-word poem
  seedpoet serve [addr]          Serve the website (default 127.0.0.1:8080)
  seedpoet version               Print the version
`, seedpoet.Version)
}
