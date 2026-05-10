package rhyme

import "sort"

type Index struct {
	Groups    map[string][]string
	ValidKeys []string // Sorted by group size (largest first)
}

// BuildPhoneticIndex creates a rhyme index from pre-computed phonetic groups.
// Keys are sorted by group size (largest first) for better variety.
func BuildPhoneticIndex() *Index {
	type keySize struct {
		key  string
		size int
	}

	var keySizes []keySize
	for key, words := range PhoneticGroups {
		if len(words) >= 2 {
			keySizes = append(keySizes, keySize{key, len(words)})
		}
	}

	// Sort by size descending - larger groups first for more variety
	sort.Slice(keySizes, func(i, j int) bool {
		return keySizes[i].size > keySizes[j].size
	})

	validKeys := make([]string, len(keySizes))
	for i, ks := range keySizes {
		validKeys[i] = ks.key
	}

	return &Index{
		Groups:    PhoneticGroups,
		ValidKeys: validKeys,
	}
}

func (idx *Index) GetGroup(key string) []string {
	return idx.Groups[key]
}

func (idx *Index) GetValidKeys() []string {
	return idx.ValidKeys
}
