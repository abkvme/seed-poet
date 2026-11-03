package server

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/abkvme/seed-poet/internal/bip39"
	"github.com/abkvme/seed-poet/internal/generator"
	"github.com/abkvme/seed-poet/internal/rhyme"
	"github.com/abkvme/seed-poet/internal/wordlist"
)

func TestPoemPage(t *testing.T) {
	gen := generator.New(wordlist.Verbs, wordlist.Adjectives, wordlist.Nouns, rhyme.BuildPhoneticIndex())
	h := New(gen).SetupRoutes()
	pre := regexp.MustCompile(`(?s)<pre id="poem-text">(.*?)</pre>`)

	for _, path := range []string{"/12", "/24"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, rec.Code)
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("%s: Cache-Control %q, want no-store", path, got)
		}
		if got := rec.Header().Get("Referrer-Policy"); got != "no-referrer" {
			t.Errorf("%s: Referrer-Policy %q", path, got)
		}
		m := pre.FindStringSubmatch(rec.Body.String())
		if m == nil {
			t.Fatalf("%s: poem not found in page", path)
		}
		words := strings.Fields(m[1])
		if err := bip39.Validate(words); err != nil {
			t.Errorf("%s: %v", path, err)
		}
	}
}
