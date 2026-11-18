package web

import (
	"testing"

	"github.com/abkvme/seed-poet/v2/data"
)

func TestWordlistHasEveryWordOnce(t *testing.T) {
	wl := BuildWordlist()
	seen := map[string]int{}
	for _, g := range append(append([]WordGroup{}, wl.Rhymes...), wl.Endings...) {
		if len(g.Words) < 2 {
			t.Errorf("group -%s has %d words", g.Key, len(g.Words))
		}
		for _, w := range g.Words {
			seen[w]++
		}
	}
	for _, w := range wl.Lone {
		seen[w]++
	}
	if wl.Total != len(data.BIP39) || len(seen) != len(data.BIP39) {
		t.Fatalf("total %d, unique %d, want %d", wl.Total, len(seen), len(data.BIP39))
	}
	for _, w := range data.BIP39 {
		if seen[w] != 1 {
			t.Errorf("%s listed %d times", w, seen[w])
		}
	}
}

func TestEnding(t *testing.T) {
	for w, want := range map[string]string{"face": "ace", "tree": "ee", "army": "y", "water": "er", "house": "ouse", "age": "age"} {
		if got := ending(w); got != want {
			t.Errorf("ending(%q) = %q, want %q", w, got, want)
		}
	}
}
