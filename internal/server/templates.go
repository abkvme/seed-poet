package server

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var (
	landingTemplate *template.Template
	poemTemplate    *template.Template
	aboutTemplate   *template.Template
	expiredTemplate *template.Template
)

func init() {
	landingTemplate = template.Must(template.ParseFS(templateFS, "templates/landing.html", "templates/strength.html"))
	poemTemplate = template.Must(template.ParseFS(templateFS, "templates/poem.html", "templates/strength.html"))
	aboutTemplate = template.Must(template.ParseFS(templateFS, "templates/about.html"))
	expiredTemplate = template.Must(template.ParseFS(templateFS, "templates/expired.html"))
}
