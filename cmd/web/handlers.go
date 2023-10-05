package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"snippetbox.lazarmrkic.com/internal/models"
	"strconv"
	"strings"
	"unicode/utf8"
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

	// ograničenje za veličinu podataka unutar forme:
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	// parsiranje forme iz "request"-a i provjera da li postoje neke greške
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// vadimo "title" i "content" iz "r.PostForm" mape
	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")
	// ukoliko trebamo da vadimo više vrijednosti odjednom (checkbox i slično), onda to možemo da odradimo preko petlje
	/*
		for i, item := range r.PostForm["items"] {
			fmt.Fprintf(w, "%d: Item %s\n", i, item)
		}
	*/

	// "r.PostForm.Get()" metoda uvijek vraća podatke iz forme u vidu String-a
	// međutim, u našem konkretnom slučaju - vrijednost za "expired" mora biti "integer" i zbog toga vršimo provjeru
	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// unutar ove mapeće biti sve potencijalne greške prilikom validacije polja
	fieldErrors := make(map[string]string)

	// odličan blog post koji sadrži generalne metode za validaciju:
	// https://www.alexedwards.net/blog/validation-snippets-for-go#email-validation
	// provjera:
	if strings.TrimSpace(title) == "" {
		fieldErrors["title"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(title) > 100 {
		fieldErrors["title"] = "This field cannot be more than 100 characters long"
	}

	if strings.TrimSpace(content) == "" {
		fieldErrors["content"] = "This field cannot be blank"
	}

	if expires != 1 && expires != 7 && expires != 365 {
		fieldErrors["expires"] = "This field must equal 1, 7 or 365"
	}

	// ukoliko neke od grešaka postoje, treba ih baciti u "plain text HTTP response" i vratiti iz handler-a
	if len(fieldErrors) > 0 {
		fmt.Fprint(w, fieldErrors)
		return
	}

	// prosljeđivanje podataka ka bazi
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// "redirect" putanja mora da se ažurira, kako bi se koristio novi, čistiji URL format
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
