package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"snippetbox.lazarmrkic.com/internal/models"
	"strconv"
)

// "home" handler će postati metoda "application" struct-a:
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// novi "httprouter" tačno ubada "/" putanju, pa zbog toga uklanjamo "r.URL.Path != "/" provjeru
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Call the newTemplateData() helper to get a templateData struct containing
	// the 'default' data (which for now is just the current year), and add the
	// snippets slice to it.
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// Pass the data to the render() helper as normal.
	app.render(w, r, http.StatusOK, "home.tmpl", data)
}

// "snippetView" handler će postati metoda "application" struct-a:
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// prilikom parsiranja request-ova u "httprouter", vrijednost svih imenovanih elemenataće biti sačuvana u "request" kontekstu
	// za sada je dovoljno da koristimo "ParamsFromContext()" funkciju da izvadimo "slice" koji sadrži nazive i vrijednosti ovih parametara
	params := httprouter.ParamsFromContext(r.Context())

	// koristićemo "params.byName()" metodu da izvadimo vrijednosti "id" imenovanog parametra iz gornjeg "slice"-a
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
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

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, r, http.StatusOK, "view.tmpl", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "create.tmpl", data)
}

// "snippetCreate" handler će postati metoda "application" struct-a:
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// sad kad smo uveli novi "router", nema potrebe da provjeravamo da li je u pitanju POST request metoda
	title := "0 snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7

	// prosljeđivanje podataka ka bazi
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// "redirect" putanja mora da se ažurira, kako bi se koristio novi, čistiji URL format
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
