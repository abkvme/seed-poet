package syllable

import "testing"

func TestCount(t *testing.T) {
	tests := []struct {
		word string
		want int
	}{
		{"apple", 2},
		{"banana", 3},
		{"cat", 1},
		{"syllable", 3},
		{"generation", 4},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			if got := Count(tt.word); got != tt.want {
				t.Errorf("Count(%q) = %d; want %d", tt.word, got, tt.want)
			}
		})
	}
}
