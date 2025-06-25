package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"snippetbox/internal/models"
	"snippetbox/ui"
	"time"
)

// A templateData type to act as the holding structure for any dynamic data passed to HTML templates.
type templateData struct {
	CurrentYear     int
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	Link            *models.LinkMapping
	Links           []*models.LinkMapping
	File            *models.File
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	User            *models.User
	BillItems       []models.BillItem
}

// A humanDate function which returns a formatted string representation of a time.Time object.
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Essentially a string-keyed map which acts as a lookup between the names of the custom template functions and the
// functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	// pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl.html")

	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Slice containing the filepath patterns for templates to be parsed
		patterns := []string{
			"html/base.tmpl.html",
			"html/partials/*.tmpl.html",
			page,
		}

		// Using ParseFS() instead of ParseFiles() to parse template files from the ui.Files EFS
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
