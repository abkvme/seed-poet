package server

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/abkvme/seed-poet/v2/seedpoet"
)

func handler(t *testing.T) http.Handler {
	t.Helper()
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	return s.Handler()
}

func do(h http.Handler, method, path string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPagesHeadersAndSEO(t *testing.T) {
	h := handler(t)
	for _, path := range []string{"/", "/about", "/expired"} {
		rec := do(h, http.MethodGet, path, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, rec.Code)
		}
		csp := rec.Header().Get("Content-Security-Policy")
		if !strings.Contains(csp, "connect-src 'none'") || strings.Contains(csp, "unsafe") {
			t.Errorf("%s: CSP %q", path, csp)
		}
		body := rec.Body.String()
		for _, want := range []string{
			`href="https://oxisoft.io"`, `<meta property="og:image" content="https://seedpoet.com/og.png">`,
			`name="twitter:card" content="summary_large_image"`, `rel="manifest"`, `rel="icon"`, `<meta name="description"`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("%s: missing %s", path, want)
			}
		}
		if regexp.MustCompile(`<style|style="|onclick=|<script>`).MatchString(body) {
			t.Errorf("%s: inline style/script would break the CSP", path)
		}
		if strings.Contains(body, `<script src=`) && !strings.Contains(body, `integrity="sha384-`) {
			t.Errorf("%s: script without integrity", path)
		}
	}
	home := do(h, http.MethodGet, "/", nil).Body.String()
	for _, want := range []string{`"@type":"WebApplication"`, `rel="canonical" href="https://seedpoet.com/"`, "Bitcoin, Ethereum, Solana", "https://github.com/abkvme/seed-poet"} {
		if !strings.Contains(home, want) {
			t.Errorf("/: missing %s", want)
		}
	}
	if !strings.Contains(do(h, http.MethodGet, "/about", nil).Body.String(), `id="blockchains"`) {
		t.Error("/about: blockchains section missing")
	}
	if r := do(h, http.MethodGet, "/expired", nil); !strings.Contains(r.Body.String(), `content="noindex`) || r.Header().Get("X-Robots-Tag") == "" {
		t.Error("/expired must be noindex")
	}
	// One generator page: old /12 and /24 links start a poem on /.
	for _, n := range []string{"12", "24"} {
		if r := do(h, http.MethodGet, "/"+n, nil); r.Code != http.StatusMovedPermanently || r.Header().Get("Location") != "/?words="+n {
			t.Errorf("/%s: %d %q", n, r.Code, r.Header().Get("Location"))
		}
	}
}

// The static generator page carries no poem: the browser writes it.
func TestPoemPagesHaveNoPoem(t *testing.T) {
	h := handler(t)
	re := regexp.MustCompile(`(?s)<ol class="poem-lines" id="poem-lines"[^>]*>(.*?)</ol>`)
	for _, path := range []string{"/"} {
		m := re.FindStringSubmatch(do(h, http.MethodGet, path, nil).Body.String())
		if m == nil || strings.TrimSpace(m[1]) != "" {
			t.Errorf("%s: poem list must be empty, got %q", path, m)
		}
	}
}

func TestNoJSFallback(t *testing.T) {
	h := handler(t)
	rec := do(h, http.MethodPost, "/nojs/24", map[string]string{"Sec-Fetch-Site": "same-origin"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != "no-store, max-age=0" {
		t.Error("nojs page must be no-store")
	}
	body := rec.Body.String()
	if !strings.Contains(body, `content="900;url=/expired"`) {
		t.Error("missing 15-minute refresh")
	}
	words := regexp.MustCompile(`(?s)<ol id="numbered-list"[^>]*>(.*?)</ol>`).FindStringSubmatch(body)
	if words == nil {
		t.Fatal("numbered list missing")
	}
	list := regexp.MustCompile(`<li>(\w+)</li>`).FindAllStringSubmatch(words[1], -1)
	var ws []string
	for _, m := range list {
		ws = append(ws, m[1])
	}
	if len(ws) != 24 {
		t.Fatalf("got %d words", len(ws))
	}
	if err := seedpoet.Validate(ws); err != nil {
		t.Error(err)
	}
	if r := do(h, http.MethodPost, "/nojs/24", map[string]string{"Sec-Fetch-Site": "cross-site"}); r.Code != http.StatusForbidden {
		t.Errorf("cross-site: %d", r.Code)
	}
	if r := do(h, http.MethodPost, "/nojs/13", nil); r.Code != http.StatusNotFound {
		t.Errorf("/nojs/13: %d", r.Code)
	}
	if r := do(h, http.MethodGet, "/nojs/12", nil); r.Code != http.StatusNotFound && r.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /nojs/12: %d", r.Code)
	}
}

func TestStaticFiles(t *testing.T) {
	h := handler(t)
	for path, ct := range map[string]string{
		"/manifest.webmanifest": "application/manifest+json", "/sw.js": "javascript", "/robots.txt": "text/plain",
		"/sitemap.xml": "xml", "/llms.txt": "text/plain", "/favicon.ico": "image", "/favicon.svg": "image/svg",
		"/og.png": "image/png", "/icons/icon-512.png": "image/png", "/seedpoet-offline.html": "text/html",
	} {
		rec := do(h, http.MethodGet, path, nil)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Content-Type"), ct) {
			t.Errorf("%s: %d %q", path, rec.Code, rec.Header().Get("Content-Type"))
		}
	}
	if !strings.Contains(do(h, http.MethodGet, "/sw.js", nil).Header().Get("Content-Security-Policy"), "connect-src 'self'") {
		t.Error("sw.js needs connect-src 'self'")
	}
	if !strings.HasPrefix(do(h, http.MethodGet, "/seedpoet-offline.html", nil).Header().Get("Content-Disposition"), "attachment") {
		t.Error("offline file must download")
	}
	for _, path := range []string{"/../go.mod", "/go.mod", "/web.go", "/%2e%2e/go.mod"} {
		if r := do(h, http.MethodGet, path, nil); r.Code == http.StatusOK {
			t.Errorf("%s: served a file outside the site", path)
		}
	}
	for path, code := range map[string]int{"/about.html": 301, "/index.html": 301, "/nope": 404, "/assets.json": 404} {
		if r := do(h, http.MethodGet, path, nil); r.Code != code {
			t.Errorf("%s: %d, want %d", path, r.Code, code)
		}
	}
}
