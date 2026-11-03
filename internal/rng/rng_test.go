package rng

import (
	"bytes"
	"testing"
)

func TestDeterministicStream(t *testing.T) {
	a, b := Deterministic([]byte("x")), Deterministic([]byte("x"))
	for i := 0; i < 200; i++ {
		if a.Uint32() != b.Uint32() {
			t.Fatal("same seed gave different streams")
		}
	}
	// First word of SHA-256("x" || 00000000) = 0x... computed independently below.
	c := Deterministic([]byte("x"))
	var stream counterStream
	stream.seed = []byte("x")
	buf := make([]byte, 4)
	stream.Read(buf)
	if got, want := c.Uint32(), uint32(buf[0])<<24|uint32(buf[1])<<16|uint32(buf[2])<<8|uint32(buf[3]); got != want {
		t.Fatalf("Uint32 = %08x, want %08x", got, want)
	}
}

func TestIntNRangeAndRejection(t *testing.T) {
	// A stream whose first value is 0xFFFFFFFF must be rejected for n = 3
	// (limit = 2^32 - 2^32 % 3 = 4294967295), then 0x00000005 -> 5 % 3 = 2.
	src := bytes.NewReader(append([]byte{0xff, 0xff, 0xff, 0xff, 0, 0, 0, 5}, make([]byte, 248)...))
	r := New(src)
	if got := r.IntN(3); got != 2 {
		t.Fatalf("IntN(3) = %d, want 2", got)
	}
	d := Deterministic([]byte("range"))
	for i := 0; i < 10000; i++ {
		n := i%97 + 1
		if v := d.IntN(n); v < 0 || v >= n {
			t.Fatalf("IntN(%d) = %d", n, v)
		}
	}
}

func TestIntNOneConsumesNothing(t *testing.T) {
	a, b := Deterministic([]byte("s")), Deterministic([]byte("s"))
	a.IntN(1)
	if a.Uint32() != b.Uint32() {
		t.Fatal("IntN(1) consumed randomness")
	}
}

func TestShuffleIsPermutation(t *testing.T) {
	d := Deterministic([]byte("shuffle"))
	xs := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	d.Shuffle(len(xs), func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
	seen := map[int]bool{}
	for _, x := range xs {
		seen[x] = true
	}
	if len(seen) != 10 {
		t.Fatalf("not a permutation: %v", xs)
	}
}
