package web

import (
	"html/template"
	"regexp"
	"sort"
	"strings"

	"github.com/abkvme/seed-poet/v2/data"
)

// WordGroup is one group on the /wordlist page.
type WordGroup struct {
	Key   string // shown as "-key"
	Label string // optional note, e.g. "unstressed"
	Words []string
}

// Wordlist is the content of the /wordlist page: every BIP39 word exactly once.
type Wordlist struct {
	Rhymes  []WordGroup // the phonetic rhyme groups Seed Poet uses to end its lines
	Endings []WordGroup // every other word, grouped by its spelled ending
	Lone    []string    // words whose spelled ending no other remaining word shares
	Ordered []Numbered  // the whole list in official order, numbered 1..2048
	Total   int
}

// Numbered is one word of the official list with its 1-based number.
type Numbered struct {
	N    int
	Word string
}

var endingRE = regexp.MustCompile(`[aeiouy]+[^aeiouy]*$`)

// ending is a word's spelled rhyme part: its last vowel group and what follows,
// with a silent final "e" kept on the consonants before it ("face" -> "ace").
func ending(w string) string {
	if strings.HasSuffix(w, "e") && len(w) > 2 {
		stem := w[:len(w)-1]
		if m := endingRE.FindString(stem); m != "" && !strings.ContainsAny(m[len(m)-1:], "aeiouy") {
			return m + "e"
		}
	}
	if m := endingRE.FindString(w); m != "" {
		return m
	}
	return w
}

// BuildWordlist groups all 2048 BIP39 words by rhyme. Output is deterministic.
func BuildWordlist() Wordlist {
	var wl Wordlist
	seen := map[string]bool{}
	for _, g := range data.Rhymes {
		wg := WordGroup{Key: g.Key}
		if k, label, ok := strings.Cut(g.Key, "_"); ok {
			wg.Key, wg.Label = k, label
		}
		if g.Key == "shun" {
			wg.Label = "tion, sion"
		}
		for _, w := range g.Words {
			if !seen[w] {
				seen[w] = true
				wg.Words = append(wg.Words, w)
			}
		}
		sort.Strings(wg.Words)
		if len(wg.Words) > 0 {
			wl.Rhymes = append(wl.Rhymes, wg)
		}
	}
	byEnd := map[string][]string{}
	for _, w := range data.BIP39 {
		if !seen[w] {
			e := ending(w)
			byEnd[e] = append(byEnd[e], w)
		}
	}
	keys := make([]string, 0, len(byEnd))
	for k := range byEnd {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		ws := byEnd[k]
		if len(ws) == 1 {
			wl.Lone = append(wl.Lone, ws[0])
			continue
		}
		wl.Endings = append(wl.Endings, WordGroup{Key: k, Words: ws})
	}
	sort.Strings(wl.Lone)
	for _, g := range wl.Rhymes {
		wl.Total += len(g.Words)
	}
	for _, g := range wl.Endings {
		wl.Total += len(g.Words)
	}
	wl.Total += len(wl.Lone)
	for i, w := range data.BIP39 {
		wl.Ordered = append(wl.Ordered, Numbered{N: i + 1, Word: w})
	}
	return wl
}

func wordlistLD() template.JS {
	return ld(map[string]any{
		"@context": "https://schema.org",
		"@graph": []any{
			map[string]any{
				"@type":       "WebPage",
				"@id":         Site + "/wordlist#page",
				"url":         Site + "/wordlist",
				"name":        "BIP39 word list, numbered and grouped by rhyme",
				"isPartOf":    map[string]string{"@id": Site + "/#website"},
				"inLanguage":  "en",
				"description": wordlistDesc,
			},
			map[string]any{
				"@type": "BreadcrumbList",
				"itemListElement": []any{
					map[string]any{"@type": "ListItem", "position": 1, "name": "Seed Poet", "item": Site + "/"},
					map[string]any{"@type": "ListItem", "position": 2, "name": "BIP39 word list", "item": Site + "/wordlist"},
				},
			},
		},
	})
}

const wordlistDesc = "All 2048 BIP39 English words, numbered 1 to 2048 in official order and grouped by " +
	"rhyme: the groups Seed Poet uses for its seed phrase poems."

var builtWordlist *Wordlist

func wordlist() *Wordlist {
	if builtWordlist == nil {
		wl := BuildWordlist()
		builtWordlist = &wl
	}
	return builtWordlist
}
