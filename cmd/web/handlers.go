package main

import (
	"errors"
	"fmt"
	"net/http"
	"snippetbox.lazarmrkic.com/internal/models"
	"strconv"
)

// "home" handler će postati metoda "application" struct-a:
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// "/" putanja ne treba da "hvata" sve zahtjeve
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	for _, snippet := range snippets {
		fmt.Fprintf(w, "+v\n", snippet)
	}

	//files := []string{
	//	"./ui/html/base.tmpl",
	//	"./ui/html/pages/home.tmpl",
	//	"./ui/html/partials/nav.tmpl",
	//}
	//
	//// "variadic" argumenti
	//ts, err := template.ParseFiles(files...)
	//if err != nil {
	//	app.serverError(w, r, err)
	//	return
	//}
	//
	//// "execute" metoda upisuje sadržaj templejta u "body" unutar odgovora
	//err = ts.ExecuteTemplate(w, "base", nil)
	//if err != nil {
	//	// opet pristupamo polju "application" struct-a:
	//	app.serverError(w, r, err)
	//	http.Error(w, "Internal server error!", http.StatusInternalServerError)
	//}
}

// "snippetView" handler će postati metoda "application" struct-a:
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// vadi se vrijednost "id" parametra iz URL-a
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	fmt.Fprintf(w, "%+v", snippet)
}

// "snippetCreate" handler će postati metoda "application" struct-a:
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// po default-u, "status code" koji se šalje je "200 OK"
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	title := "0 snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7

	// prosljeđivanje podataka ka bazi
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}
