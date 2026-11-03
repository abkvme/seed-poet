package server

import (
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/abkvme/seed-poet/internal/generator"
)

// PoemTTL is how long a generated poem stays on the user's page.
const PoemTTL = 15 * time.Minute

// contentSecurityPolicy allows nothing but this origin's own CSS, JS and API.
const contentSecurityPolicy = "default-src 'none'; script-src 'self'; style-src 'self'; " +
	"img-src 'self' data:; connect-src 'self'; form-action 'self'; base-uri 'none'; " +
	"frame-ancestors 'none'"

type Server struct {
	gen *generator.Generator
}

type PoemPageData struct {
	WordCount  int
	Poem       string // only set for the no-JavaScript fallback (POST)
	TTL        int
	TTLMinutes int
}

func New(gen *generator.Generator) *Server {
	return &Server{gen: gen}
}

func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Web routes
	mux.HandleFunc("GET /{$}", s.handleLanding)
	mux.HandleFunc("GET /12", s.handlePoemShell(12))
	mux.HandleFunc("GET /24", s.handlePoemShell(24))
	mux.HandleFunc("POST /12", s.handlePoemFallback(12))
	mux.HandleFunc("POST /24", s.handlePoemFallback(24))
	mux.HandleFunc("GET /expired", s.handleExpired)
	mux.HandleFunc("GET /about", s.handleAbout)

	// API: the only way the JavaScript page gets a poem.
	mux.HandleFunc("POST /api/poem", s.handleAPIPoem)

	// Static assets (CSS/JS), no directory listings.
	sub, _ := fs.Sub(staticFS, "static")
	files := http.StripPrefix("/static/", http.FileServerFS(sub))
	mux.Handle("GET /static/css/{file}", files)
	mux.Handle("GET /static/js/{file}", files)

	// SEO routes
	mux.HandleFunc("GET /robots.txt", s.handleStatic("static/robots.txt", "text/plain; charset=utf-8"))
	mux.HandleFunc("GET /llms.txt", s.handleStatic("static/llms.txt", "text/plain; charset=utf-8"))
	mux.HandleFunc("GET /llm.txt", s.handleStatic("static/llms.txt", "text/plain; charset=utf-8"))
	mux.HandleFunc("GET /sitemap.xml", s.handleStatic("static/sitemap.xml", "application/xml; charset=utf-8"))

	return securityHeaders(mux)
}

// securityHeaders applies to every response. Poem responses carry a seed phrase:
// nothing may be cached by a browser or proxy, framed, indexed or leak a referrer.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		if strings.HasPrefix(r.URL.Path, "/static/") {
			h.Set("Cache-Control", "public, max-age=3600")
		} else {
			h.Set("Cache-Control", "no-store, max-age=0")
			h.Set("Pragma", "no-cache")
		}
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), "+
			"magnetometer=(), microphone=(), payment=(), usb=(), interest-cohort=(), browsing-topics=()")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Run(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.SetupRoutes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	log.Printf("Starting server on %s", addr)
	return srv.ListenAndServe()
}

func render(w http.ResponseWriter, t *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

func pageData(wordCount int, poem string) PoemPageData {
	return PoemPageData{
		WordCount:  wordCount,
		Poem:       poem,
		TTL:        int(PoemTTL.Seconds()),
		TTLMinutes: int(PoemTTL.Minutes()),
	}
}

func (s *Server) handleLanding(w http.ResponseWriter, r *http.Request) {
	render(w, landingTemplate, nil)
}

// handlePoemShell serves the poem page WITHOUT a poem. The page's script asks
// /api/poem for one, so the seed never sits in a URL, in a cached document, or
// in anything a crawler or link previewer can fetch with a GET.
func (s *Server) handlePoemShell(wordCount int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
		render(w, poemTemplate, pageData(wordCount, ""))
	}
}

// handlePoemFallback serves browsers without JavaScript: a POST from the
// <noscript> form returns a page holding the poem that refreshes itself to
// /expired after PoemTTL.
func (s *Server) handlePoemFallback(wordCount int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !sameOrigin(r) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		poem, err := s.gen.Generate(wordCount)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Generate error: %v", err)
			return
		}
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
		render(w, poemTemplate, pageData(wordCount, poem.Format()))
	}
}

type apiRequest struct {
	Words int `json:"words"`
}

type apiResponse struct {
	Lines []string `json:"lines"`
	TTL   int      `json:"ttl"`
}

func (s *Server) handleAPIPoem(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) || !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var req apiRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 256)).Decode(&req); err != nil ||
		(req.Words != 12 && req.Words != 24) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	poem, err := s.gen.Generate(req.Words)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Generate error: %v", err)
		return
	}
	lines := make([]string, len(poem.Lines))
	for i, l := range poem.Lines {
		lines[i] = strings.Join(l, " ")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	json.NewEncoder(w).Encode(apiResponse{Lines: lines, TTL: int(PoemTTL.Seconds())})
}

// sameOrigin rejects requests that a browser marks as coming from another site.
// Clients that send no Sec-Fetch-Site (curl, old browsers) are allowed: they can
// only ever receive their own freshly generated poem.
func sameOrigin(r *http.Request) bool {
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "same-origin", "none":
		return true
	}
	return false
}

func (s *Server) handleExpired(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	render(w, expiredTemplate, pageData(0, ""))
}

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	render(w, aboutTemplate, nil)
}

func (s *Server) handleStatic(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFS.ReadFile(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	}
}
