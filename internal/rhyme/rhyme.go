// Package rhyme indexes the BIP39 words by rhyme group (data/seedpoet.json).
package rhyme

import (
	"sort"

	"github.com/abkvme/seed-poet/v2/data"
)

// Index maps a rhyme key to its words and lists the keys usable for couplets.
type Index struct {
	Groups    map[string][]string
	ValidKeys []string // largest group first, ties by key (same order in the JS library)
}

// BuildPhoneticIndex creates a rhyme index from the rhyme groups in data/seedpoet.json.
// Only groups with 2+ words are used. The key order is fully deterministic so that
// the Go and JavaScript generators consume randomness identically.
func BuildPhoneticIndex() *Index {
	groups := make(map[string][]string, len(data.Rhymes))
	var keys []string
	for _, g := range data.Rhymes {
		groups[g.Key] = g.Words
		if len(g.Words) >= 2 {
			keys = append(keys, g.Key)
		}
	}
	sort.Strings(keys)
	sort.SliceStable(keys, func(i, j int) bool {
		return len(groups[keys[i]]) > len(groups[keys[j]])
	})
	return &Index{Groups: groups, ValidKeys: keys}
}

// GetGroup returns the words of the rhyme group key (nil if unknown).
func (idx *Index) GetGroup(key string) []string {
	return idx.Groups[key]
}

// GetValidKeys returns the keys of groups with at least two words, largest first.
func (idx *Index) GetValidKeys() []string {
	return idx.ValidKeys
}
