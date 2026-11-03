// Package server serves the Seed Poet website for self-hosting (`seedpoet serve`)
// and, behind nginx in production, the no-JavaScript fallback (POST /nojs/12|24).
//
// Normal visitors never ask the server for a poem: their browser writes it.
package server

import (
	"bytes"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/abkvme/seed-poet/v2/seedpoet"
	"github.com/abkvme/seed-poet/v2/web"
)

// Page CSP: the page may not connect anywhere, so a poem cannot leave the device.
const pageCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data:; " +
	"font-src 'self'; manifest-src 'self'; worker-src 'self'; connect-src 'none'; form-action 'self'; " +
	"base-uri 'none'; frame-ancestors 'none'"

// The service worker needs to fetch the site's own files to cache them.
const workerCSP = "default-src 'none'; connect-src 'self'"

type Server struct {
	files    fs.FS
	renderer *web.Renderer
}

func New() (*Server, error) {
	files, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return nil, err
	}
	assets, err := web.LoadAssets()
	if err != nil {
		return nil, err
	}
	r, err := web.NewRenderer(assets, nil)
	if err != nil {
		return nil, err
	}
	return &Server{files: files, renderer: r}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /nojs/{words}", s.handleNoJS)
	mux.HandleFunc("GET /", s.handleStatic)
	return headers(mux)
}

func (s *Server) Run(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	log.Printf("Seed Poet v%s on http://%s", seedpoet.Version, addr)
	return srv.ListenAndServe()
}

func headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", pageCSP)
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

// handleStatic serves web/dist with clean URLs (/about -> about.html) and a real 404.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	p := path.Clean("/" + r.URL.Path)
	name := strings.TrimPrefix(p, "/")
	if name == "" {
		name = "index.html"
	}
	if strings.HasSuffix(p, ".html") && p != "/seedpoet-offline.html" {
		// One canonical URL per page: /about, never /about.html (and / for /index.html).
		to := strings.TrimSuffix(p, ".html")
		if to == "/index" {
			to = "/"
		}
		http.Redirect(w, r, to, http.StatusMovedPermanently)
		return
	}
	if p == "/12" || p == "/24" {
		// The generator is one page now; old links and bookmarks start a poem there.
		http.Redirect(w, r, "/?words="+name, http.StatusMovedPermanently)
		return
	}
	if name == "assets.json" {
		http.NotFound(w, r)
		return
	}
	b, err := fs.ReadFile(s.files, name)
	if err != nil {
		if b2, err2 := fs.ReadFile(s.files, name+".html"); err2 == nil && !strings.Contains(name, ".") {
			b, name = b2, name+".html"
		} else {
			s.notFound(w)
			return
		}
	}
	h := w.Header()
	switch {
	case strings.HasPrefix(name, "assets/"):
		h.Set("Cache-Control", "public, max-age=31536000, immutable")
	case name == "sw.js":
		h.Set("Cache-Control", "no-cache")
		h.Set("Content-Security-Policy", workerCSP)
	case name == "seedpoet-offline.html":
		h.Set("Content-Disposition", `attachment; filename="seedpoet-offline.html"`)
		h.Set("Cache-Control", "no-cache")
	case strings.HasSuffix(name, ".html"):
		h.Set("Cache-Control", "no-cache")
	default:
		h.Set("Cache-Control", "public, max-age=86400")
	}
	switch name {
	case "expired.html":
		h.Set("X-Robots-Tag", "noindex, follow")
	}
	if ct := contentType(name); ct != "" {
		h.Set("Content-Type", ct)
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(b))
}

func contentType(name string) string {
	switch path.Ext(name) {
	case ".webmanifest":
		return "application/manifest+json"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".ico":
		return "image/x-icon"
	}
	return ""
}

func (s *Server) notFound(w http.ResponseWriter) {
	b, err := fs.ReadFile(s.files, "404.html")
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
	w.Write(b)
}

// handleNoJS writes a poem on the server for browsers without JavaScript. The
// poem is not stored or logged; the page refreshes to /expired after 15 minutes.
func (s *Server) handleNoJS(w http.ResponseWriter, r *http.Request) {
	words := 0
	switch r.PathValue("words") {
	case "12":
		words = 12
	case "24":
		words = 24
	default:
		s.notFound(w)
		return
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "same-origin", "none":
	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	p, err := seedpoet.Generate(words)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("generate: %v", err)
		return
	}
	var buf bytes.Buffer
	if err := s.renderer.Render(&buf, web.NoJSPage(words, p.Lines)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("render: %v", err)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Cache-Control", "no-store, max-age=0")
	h.Set("Pragma", "no-cache")
	h.Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	w.Write(buf.Bytes())
}
