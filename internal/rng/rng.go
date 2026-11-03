// Package rng turns a stream of random bytes into the integers and shuffles the
// generator needs. The JavaScript library (js/seedpoet.js) implements exactly the
// same rules, so the same bytes always give the same poem in both languages:
//
//   - Uint32: the next 4 bytes of the stream, big-endian.
//   - IntN(n): 0 if n == 1 (consumes nothing); otherwise draw Uint32 values until
//     one is below the largest multiple of n that fits in 2^32, return it mod n.
//     This rejection step keeps every result exactly equally likely.
//   - Shuffle(n): Fisher-Yates, for i from n-1 down to 1 swap i with IntN(i+1).
//
// Production uses crypto/rand. Tests use Deterministic, a SHA-256 counter stream:
// block k = SHA-256(seed || uint32_be(k)), k = 0, 1, 2, ...
package rng

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
)

// RNG draws integers from a byte stream. It is not safe for concurrent use.
type RNG struct {
	src io.Reader
	buf [256]byte
	pos int
	n   int
}

// New reads randomness from src.
func New(src io.Reader) *RNG { return &RNG{src: src} }

// Crypto reads from the operating system's secure random source.
func Crypto() *RNG { return New(crand.Reader) }

// Deterministic returns a reproducible stream for tests and cross-language vectors.
// Never use it to make real seeds.
func Deterministic(seed []byte) *RNG {
	return New(&counterStream{seed: append([]byte(nil), seed...)})
}

// Uint32 returns the next 4 bytes as a big-endian integer.
func (r *RNG) Uint32() uint32 {
	var b [4]byte
	for i := range b {
		if r.pos == r.n {
			n, err := io.ReadFull(r.src, r.buf[:])
			if err != nil {
				panic("rng: random source failed: " + err.Error())
			}
			r.pos, r.n = 0, n
		}
		b[i] = r.buf[r.pos]
		r.pos++
	}
	return binary.BigEndian.Uint32(b[:])
}

// IntN returns a uniform integer in [0, n). It panics if n <= 0.
func (r *RNG) IntN(n int) int {
	if n <= 0 {
		panic("rng: IntN with n <= 0")
	}
	if n == 1 {
		return 0
	}
	const span = uint64(1) << 32
	limit := span - span%uint64(n)
	for {
		x := uint64(r.Uint32())
		if x < limit {
			return int(x % uint64(n))
		}
	}
}

// Shuffle permutes n items uniformly (Fisher-Yates, high index first).
func (r *RNG) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i > 0; i-- {
		swap(i, r.IntN(i+1))
	}
}

// Wipe clears buffered random bytes.
func (r *RNG) Wipe() {
	clear(r.buf[:])
	r.pos, r.n = 0, 0
}

type counterStream struct {
	seed  []byte
	block uint32
	buf   []byte
}

func (c *counterStream) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) {
		if len(c.buf) == 0 {
			h := sha256.New()
			h.Write(c.seed)
			var k [4]byte
			binary.BigEndian.PutUint32(k[:], c.block)
			h.Write(k[:])
			c.buf = h.Sum(nil)
			c.block++
		}
		m := copy(p[n:], c.buf)
		c.buf = c.buf[m:]
		n += m
	}
	return n, nil
}
