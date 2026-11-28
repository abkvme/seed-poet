package web

import (
	"html/template"
	"strings"
)

// Translations of the explanation pages (SP-24). The generator (/) and the word
// list stay English: the BIP39 English list is the only one wallets reliably
// import, so the poem itself is always English. Each translated page has its own
// template (about.es.html, ...) and lives under /<lang>/ with hreflang links.

// UI holds the strings of the shared header and footer.
type UI struct {
	Skip, Generator, About, Wordlist, Privacy, WordlistLong, Memorizing, Source, Offline, Credit, Language string
}

// Alt is one language version of a page.
type Alt struct {
	Lang string // hreflang, e.g. "pt"
	Name string // language name in that language
	Path string // URL path
	URL  string // absolute URL
}

type language struct {
	Code   string // URL prefix and hreflang
	HTML   string // <html lang>
	Locale string // og:locale
	Name   string
	UI     UI
}

var english = language{Code: "en", HTML: "en", Locale: "en_US", Name: "English", UI: UI{
	Skip: "Skip to content", Generator: "Generator", About: "About", Wordlist: "Wordlist",
	Privacy: "Privacy", WordlistLong: "BIP39 word list", Memorizing: "Memorizing", Source: "Source code",
	Offline: "Offline version", Credit: "Developed by", Language: "Language",
}}

// Translations, in the order they appear in the language links.
var translations = []language{
	{Code: "es", HTML: "es", Locale: "es_ES", Name: "Español", UI: UI{
		Skip: "Saltar al contenido", Generator: "Generador", About: "Acerca de", Wordlist: "Lista de palabras",
		Privacy: "Privacidad", WordlistLong: "Lista de palabras BIP39", Memorizing: "Memorizar", Source: "Código fuente",
		Offline: "Versión sin conexión", Credit: "Desarrollado por", Language: "Idioma",
	}},
	{Code: "pt", HTML: "pt-BR", Locale: "pt_BR", Name: "Português", UI: UI{
		Skip: "Pular para o conteúdo", Generator: "Gerador", About: "Sobre", Wordlist: "Lista de palavras",
		Privacy: "Privacidade", WordlistLong: "Lista de palavras BIP39", Memorizing: "Memorizar", Source: "Código-fonte",
		Offline: "Versão offline", Credit: "Desenvolvido por", Language: "Idioma",
	}},
}

// pageText is the per-language metadata of one translated page.
type pageText struct {
	Title, Description, OGTitle, Crumb string
}

// translatedPage describes an English page and its translations.
type translatedPage struct {
	Template string // English template, e.g. "about.html"; translations use "about.<lang>.html"
	Path     string // English URL path
	Text     map[string]pageText
	LD       func(lang, path string, t pageText) template.JS
}

var translatedPages = []translatedPage{
	{Template: "about.html", Path: "/about", LD: aboutLD, Text: map[string]pageText{
		"en": {Title: "About Seed Poet: How Rhyming BIP39 Seed Phrases Work",
			Description: "How Seed Poet writes a BIP39 seed phrase as a rhyming poem, which blockchains and wallets accept it, how strong it is, and how it keeps it private.",
			OGTitle:     "About Seed Poet", Crumb: "About"},
		"es": {Title: "Acerca de Seed Poet: frases semilla BIP39 en verso",
			Description: "Cómo Seed Poet escribe una frase semilla BIP39 como un poema con rima, qué blockchains y carteras la aceptan, qué tan segura es y cómo protege tu privacidad.",
			OGTitle:     "Acerca de Seed Poet", Crumb: "Acerca de"},
		"pt": {Title: "Sobre o Seed Poet: frases-semente BIP39 em versos",
			Description: "Como o Seed Poet escreve uma frase-semente BIP39 como um poema com rima, quais blockchains e carteiras a aceitam, qual a sua força e como protege a sua privacidade.",
			OGTitle:     "Sobre o Seed Poet", Crumb: "Sobre"},
	}},
	{Template: "guide-memorize.html", Path: guideMemorizePath, LD: guideMemorizeLD, Text: map[string]pageText{
		"en": {Title: "How to Memorize a Seed Phrase, and When Not To | Seed Poet",
			Description: guideMemorizeDesc, OGTitle: "How to memorize a seed phrase, and when not to", Crumb: "Memorize a seed phrase"},
		"es": {Title: "Cómo memorizar una frase semilla, y cuándo no hacerlo | Seed Poet",
			Description: "Cómo aprender de memoria una frase semilla BIP39 de 12 o 24 palabras con rima, ejercicios y repetición espaciada, y por qué la memoria debe ser una segunda copia, nunca la única.",
			OGTitle:     "Cómo memorizar una frase semilla, y cuándo no hacerlo", Crumb: "Memorizar una frase semilla"},
		"pt": {Title: "Como memorizar uma frase-semente, e quando não fazer isso | Seed Poet",
			Description: "Como decorar uma frase-semente BIP39 de 12 ou 24 palavras com rima, exercícios e repetição espaçada, e por que a memória deve ser uma segunda cópia, nunca a única.",
			OGTitle:     "Como memorizar uma frase-semente, e quando não fazer isso", Crumb: "Memorizar uma frase-semente"},
	}},
}

// localized returns the English page and its translations, all linked by hreflang.
func (tp translatedPage) localized() []Page {
	all := append([]language{english}, translations...)
	var alts []Alt
	for _, l := range all {
		p := tp.Path
		if l.Code != "en" {
			p = "/" + l.Code + tp.Path
		}
		alts = append(alts, Alt{Lang: l.Code, Name: l.Name, Path: p, URL: Site + p})
	}
	var pages []Page
	for i, l := range all {
		t := tp.Text[l.Code]
		tmpl, root := tp.Template, ""
		if l.Code != "en" {
			tmpl = strings.TrimSuffix(tp.Template, ".html") + "." + l.Code + ".html"
			root = "/" + l.Code
		}
		p := base(Page{Template: tmpl, File: strings.TrimPrefix(alts[i].Path, "/") + ".html", Path: alts[i].Path,
			Key: tp.Path, Title: t.Title, Description: t.Description, OGTitle: t.OGTitle, OGDescription: t.Description,
			Canonical: alts[i].URL, JSONLD: tp.LD(l.Code, alts[i].Path, t), InSitemap: true,
			Lang: l.HTML, Locale: l.Locale, UI: l.UI, Root: root, Alternates: alts})
		if l.Code == "en" {
			p.OGDescription = ogDesc
			if tp.Path == guideMemorizePath {
				p.OGDescription = t.Description
			}
		}
		pages = append(pages, p)
	}
	return pages
}

func breadcrumb(lang, path, crumb string) map[string]any {
	return map[string]any{
		"@type": "BreadcrumbList",
		"itemListElement": []any{
			map[string]any{"@type": "ListItem", "position": 1, "name": "Seed Poet", "item": Site + "/"},
			map[string]any{"@type": "ListItem", "position": 2, "name": crumb, "item": Site + path},
		},
	}
}
