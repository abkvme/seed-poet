package web

import "html/template"

const guideMemorizePath = "/guide/memorize-seed-phrase"

const guideMemorizeDesc = "How to learn a 12 or 24-word BIP39 seed phrase by heart with rhyme, drills and spaced " +
	"repetition, and why memory should be a second copy, never the only backup."

func guideMemorizeLD() template.JS {
	return ld(map[string]any{
		"@context": "https://schema.org",
		"@graph": []any{
			map[string]any{
				"@type":         "Article",
				"@id":           Site + guideMemorizePath + "#article",
				"url":           Site + guideMemorizePath,
				"headline":      "How to memorize a seed phrase, and when not to",
				"description":   guideMemorizeDesc,
				"inLanguage":    "en",
				"isPartOf":      map[string]string{"@id": Site + "/#website"},
				"about":         map[string]string{"@id": Site + "/#app"},
				"author":        map[string]string{"@id": "https://oxisoft.io/#organization"},
				"publisher":     map[string]string{"@id": "https://oxisoft.io/#organization"},
				"datePublished": "2026-11-19",
				"image":         Site + "/og.png",
			},
			org,
			map[string]any{
				"@type": "BreadcrumbList",
				"itemListElement": []any{
					map[string]any{"@type": "ListItem", "position": 1, "name": "Seed Poet", "item": Site + "/"},
					map[string]any{"@type": "ListItem", "position": 2, "name": "Memorize a seed phrase", "item": Site + guideMemorizePath},
				},
			},
		},
	})
}
