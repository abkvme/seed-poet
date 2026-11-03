// Package web holds the Seed Poet website: HTML templates, static sources, and
// the generated site in dist/ (built by `go run ./cmd/buildweb`). The same
// templates render the static pages and the no-JavaScript server page, so the
// two never drift apart.
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"

	"github.com/abkvme/seed-poet/v2/seedpoet"
)

//go:embed templates/*.html
var templateFS embed.FS

// Dist is the built website (static files served as is).
//
//go:embed all:dist
var Dist embed.FS

const (
	Site       = "https://seedpoet.com"
	TTLSeconds = 900
	Repo       = "https://github.com/abkvme/seed-poet"
)

// Asset is one fingerprinted file in dist/assets.
type Asset struct {
	Path string `json:"path"` // URL path, e.g. /assets/site.1a2b3c4d5e.css
	SRI  string `json:"sri"`  // integrity value, sha384-...
}

// Page is everything a template needs.
type Page struct {
	Template      string
	File          string // output file in dist, e.g. "12.html"
	Path          string // URL path, e.g. "/12"
	Title         string
	Description   string
	OGTitle       string
	OGDescription string
	Robots        string
	Canonical     string
	URL           string
	Site          string
	Base          string // "" on the site, absolute URL in the offline file
	Words         int
	Other         int
	Poem          [][]string
	TTL           int
	TTLMinutes    int
	Offline       bool
	Version       string
	JSONLD        template.JS
	InSitemap     bool
}

// Renderer renders pages with fingerprinted asset URLs (and, for the offline
// file, inlined sources).
type Renderer struct {
	t *template.Template
}

// NewRenderer builds the templates. assets maps a source name ("css/site.css")
// to its fingerprinted path; inline returns the raw source (nil outside the build).
func NewRenderer(assets map[string]Asset, inline func(name string) (string, error)) (*Renderer, error) {
	funcs := template.FuncMap{
		"asset": func(name string) (string, error) {
			a, ok := assets[name]
			if !ok {
				return "", fmt.Errorf("unknown asset %q", name)
			}
			return a.Path, nil
		},
		"sri": func(name string) (string, error) {
			a, ok := assets[name]
			if !ok {
				return "", fmt.Errorf("unknown asset %q", name)
			}
			return a.SRI, nil
		},
		"inlineCSS": func(name string) (template.CSS, error) {
			if inline == nil {
				return "", fmt.Errorf("inline not available")
			}
			s, err := inline(name)
			return template.CSS(s), err
		},
		"inlineJS": func(name string) (template.JS, error) {
			if inline == nil {
				return "", fmt.Errorf("inline not available")
			}
			s, err := inline(name)
			// Keep "</script" from closing the element early.
			return template.JS(strings.ReplaceAll(s, "</script", "<\\/script")), err
		},
		"lineNo": func(i int) string { return fmt.Sprintf("%02d", i+1) },
	}
	t, err := template.New("").Funcs(funcs).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{t: t}, nil
}

// Render writes page p.
func (r *Renderer) Render(w io.Writer, p Page) error {
	return r.t.ExecuteTemplate(w, p.Template, p)
}

// LoadAssets reads dist/assets.json (written by buildweb).
func LoadAssets() (map[string]Asset, error) {
	b, err := fs.ReadFile(Dist, "dist/assets.json")
	if err != nil {
		return nil, err
	}
	m := map[string]Asset{}
	return m, json.Unmarshal(b, &m)
}

func base(p Page) Page {
	p.Site = Site
	p.URL = Site + p.Path
	if p.Path == "/" {
		p.URL = Site + "/"
	}
	p.TTL = TTLSeconds
	p.TTLMinutes = TTLSeconds / 60
	p.Version = seedpoet.Version
	if p.Robots == "" {
		p.Robots = "index, follow, max-image-preview:large"
	}
	if p.OGTitle == "" {
		p.OGTitle = p.Title
	}
	if p.OGDescription == "" {
		p.OGDescription = p.Description
	}
	return p
}

const (
	homeTitle = "Seed Poet: Your BIP39 Seed Phrase as a Rhyming Poem"
	homeDesc  = "A valid 12 or 24-word BIP39 seed phrase for Bitcoin, Ethereum, Solana or Cardano, written as a " +
		"rhyming poem. Made in your browser, never sent or stored."
	ogTitle = "Seed Poet: your keys, in verse"
	ogDesc  = "A valid BIP39 seed phrase as a rhyming poem. Written by your browser, never sent anywhere, gone in 15 minutes."
)

func ld(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return template.JS(b)
}

var org = map[string]any{
	"@type": "Organization",
	"@id":   "https://oxisoft.io/#organization",
	"name":  "OxiSoft",
	"url":   "https://oxisoft.io",
}

func homeLD() template.JS {
	return ld(map[string]any{
		"@context": "https://schema.org",
		"@graph": []any{
			map[string]any{
				"@type":       "WebSite",
				"@id":         Site + "/#website",
				"url":         Site + "/",
				"name":        "Seed Poet",
				"description": homeDesc,
				"inLanguage":  "en",
				"publisher":   map[string]string{"@id": "https://oxisoft.io/#organization"},
			},
			org,
			map[string]any{
				"@type":                  "WebApplication",
				"@id":                    Site + "/#app",
				"name":                   "Seed Poet",
				"url":                    Site + "/",
				"description":            homeDesc,
				"applicationCategory":    "SecurityApplication",
				"applicationSubCategory": "BIP39 seed phrase generator",
				"operatingSystem":        "Any (web browser)",
				"browserRequirements":    "Requires JavaScript; a server fallback works without it",
				"softwareVersion":        seedpoet.Version,
				"isAccessibleForFree":    true,
				"offers":                 map[string]any{"@type": "Offer", "price": "0", "priceCurrency": "USD"},
				"featureList": []string{
					"Valid BIP39 mnemonic with checksum, 12 or 24 words",
					"Rhyming couplets for easier memorising",
					"Generated on your device with a secure random source",
					"Nothing sent or stored; auto-hide after 15 minutes",
					"Printable PDF made in the browser",
					"Works offline and installs as an app",
					"Works with Bitcoin, Ethereum and EVM chains, Solana, Cardano, Polkadot, Cosmos, Tron, Litecoin, Dogecoin and any BIP39 wallet",
				},
				"keywords": "BIP39 generator, seed phrase generator, mnemonic generator, recovery phrase, Bitcoin seed phrase, " +
					"Ethereum seed phrase, Solana seed phrase, Cardano seed phrase, crypto wallet, rhyming seed phrase, offline seed generator",
				"image":      Site + "/og.png",
				"screenshot": Site + "/screenshots/wide.png",
				"author":     map[string]string{"@id": "https://oxisoft.io/#organization"},
				"publisher":  map[string]string{"@id": "https://oxisoft.io/#organization"},
				"sameAs":     []string{Repo},
			},
			map[string]any{
				"@type":               "SoftwareSourceCode",
				"@id":                 Repo,
				"name":                "seed-poet",
				"codeRepository":      Repo,
				"programmingLanguage": []string{"Go", "JavaScript"},
				"license":             "https://opensource.org/licenses/MIT",
				"author":              map[string]string{"@id": "https://oxisoft.io/#organization"},
			},
		},
	})
}

func aboutLD() template.JS {
	return ld(map[string]any{
		"@context": "https://schema.org",
		"@graph": []any{
			map[string]any{
				"@type":       "AboutPage",
				"@id":         Site + "/about#page",
				"url":         Site + "/about",
				"name":        "About Seed Poet",
				"isPartOf":    map[string]string{"@id": Site + "/#website"},
				"about":       map[string]string{"@id": Site + "/#app"},
				"inLanguage":  "en",
				"description": "How Seed Poet turns a BIP39 seed phrase into a rhyming poem, how strong it is, and how it protects your privacy.",
			},
			map[string]any{
				"@type": "BreadcrumbList",
				"itemListElement": []any{
					map[string]any{"@type": "ListItem", "position": 1, "name": "Seed Poet", "item": Site + "/"},
					map[string]any{"@type": "ListItem", "position": 2, "name": "About", "item": Site + "/about"},
				},
			},
		},
	})
}

// Pages lists every static page of the site. The generator is one page: "/".
func Pages() []Page {
	return []Page{
		base(Page{Template: "index.html", File: "index.html", Path: "/", Title: homeTitle, Description: homeDesc,
			OGTitle: ogTitle, OGDescription: ogDesc, Canonical: Site + "/", JSONLD: homeLD(), InSitemap: true}),
		base(Page{Template: "about.html", File: "about.html", Path: "/about",
			Title:       "About Seed Poet: How Rhyming BIP39 Seed Phrases Work",
			Description: "How Seed Poet writes a BIP39 seed phrase as a rhyming poem, which blockchains and wallets accept it, how strong it is, and how it keeps it private.",
			OGTitle:     "About Seed Poet", OGDescription: ogDesc,
			Canonical: Site + "/about", JSONLD: aboutLD(), InSitemap: true}),
		base(Page{Template: "expired.html", File: "expired.html", Path: "/expired",
			Title: "Poem expired | Seed Poet", Description: "This poem was removed for your safety. Generate a new one.",
			OGTitle: ogTitle, OGDescription: ogDesc, Robots: "noindex, follow"}),
		base(Page{Template: "404.html", File: "404.html", Path: "/404",
			Title: "Page not found | Seed Poet", Description: "This page does not exist.",
			OGTitle: ogTitle, OGDescription: ogDesc, Robots: "noindex, follow"}),
		base(Page{Template: "offline.html", File: "seedpoet-offline.html", Path: "/seedpoet-offline.html",
			Title:       "Seed Poet offline",
			Description: "Seed Poet in one file: generate a rhyming BIP39 seed phrase on a computer without internet.",
			OGTitle:     ogTitle, OGDescription: ogDesc, Robots: "noindex, nofollow",
			Offline: true, Base: Site}),
	}
}

// NoJSPage is the server-rendered poem page for browsers without JavaScript.
func NoJSPage(words int, lines [][]string) Page {
	return base(Page{Template: "poem.html", Path: fmt.Sprintf("/nojs/%d", words),
		Title:       fmt.Sprintf("%d-word poem | Seed Poet", words),
		Description: homeDesc, OGTitle: ogTitle, OGDescription: ogDesc,
		Robots: "noindex, nofollow, noarchive", Words: words, Poem: lines})
}
