package syllable

import "testing"

func TestCount(t *testing.T) {
	tests := []struct {
		word string
		want int
	}{
		{"", 0},
		{"apple", 2},
		{"banana", 3},
		{"cat", 1},
		{"CAT", 1},
		{"syllable", 3},
		{"generation", 4},
		{"make", 1},   // silent e
		{"the", 1},    // too short for the silent-e rule
		{"jumped", 1}, // silent -ed
		{"wanted", 2}, // -ted keeps its syllable
		{"added", 2},  // -ded keeps its syllable
		{"yard", 1},   // leading y is a consonant
		{"y", 1},      // never below 1
		{"rhythm", 1},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			if got := Count(tt.word); got != tt.want {
				t.Errorf("Count(%q) = %d; want %d", tt.word, got, tt.want)
			}
		})
	}
}

func TestCountAll(t *testing.T) {
	got := CountAll([]string{"cat", "banana", "cat"})
	if len(got) != 2 || got["cat"] != 1 || got["banana"] != 3 {
		t.Errorf("CountAll = %v", got)
	}
}
