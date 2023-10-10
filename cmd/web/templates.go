package main

import (
	"html/template"
	"path/filepath"
	"snippetbox.lazarmrkic.com/internal/models"
	"time"
)

// "templateData" će biti "struct" koji sadrži sve dinamičke podatke koje prosljeđujemo ka HTML templejtima
type templateData struct {
	CurrentYear int
	Snippet     models.Snippet
	Snippets    []models.Snippet
	Form        any
	Flash       string
	// na osnovu ovog polja će se prikazivati odgovarajući Login screen ("Authenticated" / "Not authenticated")
	IsAuthenticated bool
	// kako bi slanje formi bilo zaštićeno od CSRF, potrebna nam je "noSurf.Token()" metoda
	// ona uzima "CSRF token" i dodaje ga u skriveno "csrf_token" polje unutar svake naše forme
	// na kraju ćemo morati da "štelujemo" ovaj atribut u svim HTML poljima gdje je navedena funkcionalnost potrebna
	// prvi korak je da dodamo "CSRFToken" polje u "templateData" struct
	CSRFToken string
}

// Create a humanDate function which returns a nicely formatted string
// representation of a time.Time object.
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable. This is
// essentially a string-keyed map which acts as a lookup between the names of our
// custom template functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// The template.FuncMap must be registered with the template set before you
		// call the ParseFiles() method. This means we have to use template.New() to
		// create an empty template set, use the Funcs() method to register the
		// template.FuncMap, and then parse the file as normal.
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
