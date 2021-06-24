package server

import (
	"embed"
	"github.com/patrick246/shortlink/pkg/persistence"
	"html/template"
	"io/fs"
	"strings"
	"time"
)

//go:embed static/*
var staticContent embed.FS

//go:embed templates/*
var templateContent embed.FS

type listTemplateData struct {
	Shortlinks []persistence.Shortlink
	Page       int64
	Total      int64
	Size       int64
	CSRF       string
}

type editTemplateData struct {
	Code string
	URL  string
	CSRF string
	TTL  time.Time
}

type pagination struct {
	Prev, Next bool
	Pages      []int64
}

var templates = make(map[string]*template.Template)

func init() {
	bases, err := fs.Glob(templateContent, "**/*.base.gohtml")
	if err != nil {
		panic(err)
	}

	pages, err := fs.Glob(templateContent, "**/*.page.gohtml")
	if err != nil {
		panic(err)
	}

	for _, page := range pages {
		content, err := templateContent.ReadFile(page)
		if err != nil {
			panic(err)
		}

		tmpl := template.New(page).Funcs(map[string]interface{}{
			"pagination": func(page, total, size int64) pagination {
				result := pagination{}

				lastPage := (total-1)/size + 1
				result.Prev = page != 0
				result.Next = page != lastPage-1

				for i := page - 3; i < page+3; i++ {
					if i < 0 || i >= lastPage {
						continue
					}
					result.Pages = append(result.Pages, i)
				}
				return result
			},
			"sub": func(a, b int64) int64 {
				return a - b
			},
			"add": func(a, b int64) int64 {
				return a + b
			},
		})

		template.Must(tmpl.Parse(string(content)))
		for _, base := range bases {
			content, err := templateContent.ReadFile(base)
			if err != nil {
				panic(err)
			}
			template.Must(tmpl.Parse(string(content)))
		}

		stripped := strings.TrimPrefix(page, "templates/")
		templates[stripped] = tmpl
	}
}
