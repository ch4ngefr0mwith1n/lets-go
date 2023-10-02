package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// "home" handler će postati metoda "application" struct-a:
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// "/" putanja ne treba da "hvata" sve zahtjeve
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/pages/home.tmpl",
		"./ui/html/partials/nav.tmpl",
	}

	// "variadic" argumenti
	ts, err := template.ParseFiles(files...)
	if err != nil {
		// pristupanje polju "application" struct-a:
		app.logger.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
		http.Error(w, "Internal server error!", http.StatusInternalServerError)
		return
	}

	// "execute" metoda upisuje sadržaj templejta u "body" unutar odgovora
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		// opet pristupamo polju "application" struct-a:
		app.logger.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
		http.Error(w, "Internal server error!", http.StatusInternalServerError)
	}
}

// "snippetView" handler će postati metoda "application" struct-a:
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// vadi se vrijednost "id" parametra iz URL-a
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
	}

	// ova funkcija prima "io.writer" interfejs
	// "http.ResponseWriter" zadovoljava uslove ovog interfejsa
	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

// "snippetCreate" handler će postati metoda "application" struct-a:
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// po default-u, "status code" koji se šalje je "200 OK"
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		//http.Error(w, "Method not allowed", 405)

		//w.WriteHeader(405)
		//w.Write([]byte("Method not allowed!"))
		return
	}

	w.Write([]byte("Creating a new snippet..."))
}
