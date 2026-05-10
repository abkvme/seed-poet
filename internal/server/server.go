package server

import (
	"log"
	"net/http"

	"github.com/abkvme/seed-poet/internal/generator"
)

type Server struct {
	gen *generator.Generator
}

type PoemPageData struct {
	WordCount int
	Poem      string
}

func New(gen *generator.Generator) *Server {
	return &Server{gen: gen}
}

func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Web routes
	mux.HandleFunc("/", s.handleLanding)
	mux.HandleFunc("/12", s.handlePoem12Web)
	mux.HandleFunc("/24", s.handlePoem24Web)
	mux.HandleFunc("/about", s.handleAbout)

	// SEO routes
	mux.HandleFunc("/robots.txt", s.handleRobots)
	mux.HandleFunc("/llms.txt", s.handleLLMs)
	mux.HandleFunc("/llm.txt", s.handleLLMs)
	mux.HandleFunc("/sitemap.xml", s.handleSitemap)

	return mux
}

func (s *Server) Run(addr string) error {
	handler := s.SetupRoutes()
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, handler)
}

func (s *Server) handleLanding(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := landingTemplate.Execute(w, nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

func (s *Server) handlePoem12Web(w http.ResponseWriter, r *http.Request) {
	s.handlePoemWeb(w, r, 12)
}

func (s *Server) handlePoem24Web(w http.ResponseWriter, r *http.Request) {
	s.handlePoemWeb(w, r, 24)
}

func (s *Server) handlePoemWeb(w http.ResponseWriter, r *http.Request, wordCount int) {
	poem := s.gen.Generate(wordCount)

	data := PoemPageData{
		WordCount: wordCount,
		Poem:      poem.Format(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := poemTemplate.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := aboutTemplate.Execute(w, nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

func (s *Server) handleRobots(w http.ResponseWriter, r *http.Request) {
	data, _ := staticFS.ReadFile("static/robots.txt")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

func (s *Server) handleLLMs(w http.ResponseWriter, r *http.Request) {
	data, _ := staticFS.ReadFile("static/llms.txt")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

func (s *Server) handleSitemap(w http.ResponseWriter, r *http.Request) {
	data, _ := staticFS.ReadFile("static/sitemap.xml")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write(data)
}
