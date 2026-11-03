package generator

import (
	"sync"
	"testing"

	"github.com/abkvme/seed-poet/v2/internal/bip39"
	"github.com/abkvme/seed-poet/v2/internal/rhyme"
	"github.com/abkvme/seed-poet/v2/internal/wordlist"
)

func newGen() *Generator {
	return New(wordlist.Verbs, wordlist.Adjectives, wordlist.Nouns, rhyme.BuildPhoneticIndex())
}

func TestGenerateValidBIP39(t *testing.T) {
	g := newGen()
	idx := rhyme.BuildPhoneticIndex()
	// A word can belong to several rhyme groups.
	rhymeKeys := map[string]map[string]bool{}
	for k, words := range idx.Groups {
		for _, w := range words {
			if rhymeKeys[w] == nil {
				rhymeKeys[w] = map[string]bool{}
			}
			rhymeKeys[w][k] = true
		}
	}
	rhymes := func(a, b string) bool {
		for k := range rhymeKeys[a] {
			if rhymeKeys[b][k] {
				return true
			}
		}
		return false
	}
	for _, n := range []int{12, 24} {
		seen := map[string]bool{}
		for i := 0; i < 300; i++ {
			p, err := g.Generate(n)
			if err != nil {
				t.Fatalf("%d words: %v", n, err)
			}
			words := p.Words()
			if len(words) != n || len(p.Lines) != n/3 {
				t.Fatalf("%d words: got %d words in %d lines", n, len(words), len(p.Lines))
			}
			if err := bip39.Validate(words); err != nil {
				t.Fatalf("%d words: %v\n%s", n, err, p.Format())
			}
			for l := 0; l+1 < len(p.Lines); l += 2 {
				a, b := p.Lines[l][2], p.Lines[l+1][2]
				if !rhymes(a, b) {
					t.Errorf("couplet does not rhyme: %q / %q", a, b)
				}
			}
			s := p.Format()
			if seen[s] {
				t.Fatalf("duplicate poem after %d draws:\n%s", i, s)
			}
			seen[s] = true
		}
	}
}

func TestUnsupportedCountFallsBackTo12(t *testing.T) {
	p, err := newGen().Generate(7)
	if err != nil {
		t.Fatal(err)
	}
	if p.WordCount != 12 || len(p.Words()) != 12 {
		t.Fatalf("got %d words", len(p.Words()))
	}
}

func TestGenerateDoesNotMutateIndex(t *testing.T) {
	g := newGen()
	before := append([]string(nil), g.rhymeIndex.GetValidKeys()...)
	for i := 0; i < 20; i++ {
		if _, err := g.Generate(24); err != nil {
			t.Fatal(err)
		}
	}
	after := g.rhymeIndex.GetValidKeys()
	for i := range before {
		if before[i] != after[i] {
			t.Fatal("Generate reordered the shared rhyme index")
		}
	}
}

// Run with -race: the web server calls Generate from many goroutines.
func TestGenerateConcurrent(t *testing.T) {
	g := newGen()
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				p, err := g.Generate(24)
				if err != nil {
					t.Error(err)
					return
				}
				if err := bip39.Validate(p.Words()); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func BenchmarkGenerate24(b *testing.B) {
	g := newGen()
	for i := 0; i < b.N; i++ {
		if _, err := g.Generate(24); err != nil {
			b.Fatal(err)
		}
	}
}
