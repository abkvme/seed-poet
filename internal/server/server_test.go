package server

import (
	"encoding/json"
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

func handler() http.Handler {
	gen := generator.New(wordlist.Verbs, wordlist.Adjectives, wordlist.Nouns, rhyme.BuildPhoneticIndex())
	return New(gen).SetupRoutes()
}

func do(h http.Handler, method, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func checkPrivateHeaders(t *testing.T, path string, rec *httptest.ResponseRecorder) {
	t.Helper()
	want := map[string]string{
		"Cache-Control":   "no-store, max-age=0",
		"Referrer-Policy": "no-referrer",
		"X-Frame-Options": "DENY",
		"X-Robots-Tag":    "noindex, nofollow, noarchive",
	}
	for k, v := range want {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("%s: %s = %q, want %q", path, k, got, v)
		}
	}
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'none'") || strings.Contains(csp, "unsafe") {
		t.Errorf("%s: weak CSP %q", path, csp)
	}
}

// A GET (what crawlers, link previews and history use) must never carry a seed.
func TestPoemShellHasNoSeed(t *testing.T) {
	h := handler()
	pre := regexp.MustCompile(`(?s)<pre id="poem-text"[^>]*>(.*?)</pre>`)
	for _, path := range []string{"/12", "/24"} {
		rec := do(h, http.MethodGet, path, "", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, rec.Code)
		}
		checkPrivateHeaders(t, path, rec)
		m := pre.FindStringSubmatch(rec.Body.String())
		if m == nil {
			t.Fatalf("%s: poem element missing", path)
		}
		if strings.TrimSpace(m[1]) != "" {
			t.Errorf("%s: GET page contains words %q", path, m[1])
		}
		if strings.Contains(rec.Body.String(), "http-equiv=\"refresh\"") {
			t.Errorf("%s: shell should not auto-refresh", path)
		}
	}
}

func TestAPIPoem(t *testing.T) {
	h := handler()
	for _, n := range []int{12, 24} {
		body := `{"words":` + map[int]string{12: "12", 24: "24"}[n] + `}`
		rec := do(h, http.MethodPost, "/api/poem", body, map[string]string{
			"Content-Type": "application/json", "Sec-Fetch-Site": "same-origin"})
		if rec.Code != http.StatusOK {
			t.Fatalf("%d: status %d %s", n, rec.Code, rec.Body)
		}
		checkPrivateHeaders(t, "/api/poem", rec)
		var resp apiResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.TTL != 900 {
			t.Errorf("ttl %d, want 900", resp.TTL)
		}
		if err := bip39.Validate(strings.Fields(strings.Join(resp.Lines, " "))); err != nil {
			t.Errorf("%d words: %v", n, err)
		}
	}
}

func TestAPIRejects(t *testing.T) {
	h := handler()
	cases := []struct {
		name, method, body string
		hdr                map[string]string
		want               int
	}{
		{"GET", http.MethodGet, "", nil, http.StatusMethodNotAllowed},
		{"cross-site", http.MethodPost, `{"words":12}`, map[string]string{"Content-Type": "application/json", "Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"form post", http.MethodPost, `{"words":12}`, map[string]string{"Content-Type": "text/plain"}, http.StatusForbidden},
		{"bad count", http.MethodPost, `{"words":13}`, map[string]string{"Content-Type": "application/json"}, http.StatusBadRequest},
		{"junk", http.MethodPost, `nope`, map[string]string{"Content-Type": "application/json"}, http.StatusBadRequest},
	}
	for _, c := range cases {
		rec := do(h, c.method, "/api/poem", c.body, c.hdr)
		if rec.Code != c.want {
			t.Errorf("%s: status %d, want %d", c.name, rec.Code, c.want)
		}
	}
}

// Without JavaScript: POST returns the poem plus a refresh to /expired after 15 minutes.
func TestNoScriptFallback(t *testing.T) {
	h := handler()
	pre := regexp.MustCompile(`(?s)<pre id="poem-text"[^>]*>(.*?)</pre>`)
	rec := do(h, http.MethodPost, "/24", "", map[string]string{"Sec-Fetch-Site": "same-origin"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	checkPrivateHeaders(t, "POST /24", rec)
	body := rec.Body.String()
	if !strings.Contains(body, `content="900;url=/expired"`) {
		t.Error("fallback page lacks the 15-minute refresh")
	}
	m := pre.FindStringSubmatch(body)
	if m == nil {
		t.Fatal("poem missing")
	}
	if err := bip39.Validate(strings.Fields(m[1])); err != nil {
		t.Error(err)
	}
	if rec := do(h, http.MethodPost, "/24", "", map[string]string{"Sec-Fetch-Site": "cross-site"}); rec.Code != http.StatusForbidden {
		t.Errorf("cross-site fallback POST: status %d", rec.Code)
	}
}

func TestPagesAndCredit(t *testing.T) {
	h := handler()
	for _, path := range []string{"/", "/about", "/12", "/expired"} {
		rec := do(h, http.MethodGet, path, "", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `href="https://oxisoft.io"`) || !strings.Contains(body, "OxiSoft") {
			t.Errorf("%s: OxiSoft credit missing", path)
		}
		if regexp.MustCompile(`<style|style="|onclick=|<script>`).MatchString(body) {
			t.Errorf("%s: inline style/script would break the CSP", path)
		}
	}
	for _, path := range []string{"/static/css/poem.css", "/static/js/poem.js", "/robots.txt"} {
		if rec := do(h, http.MethodGet, path, "", nil); rec.Code != http.StatusOK {
			t.Errorf("%s: status %d", path, rec.Code)
		}
	}
	for _, path := range []string{"/static/", "/static/css/", "/nope"} {
		if rec := do(h, http.MethodGet, path, "", nil); rec.Code != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", path, rec.Code)
		}
	}
}
