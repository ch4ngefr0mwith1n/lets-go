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

// ovaj "struct" predstavlja podatke unutar forme i greške tokom validacije za njena polja
// sva njegova polja su u Pascal case, zato što moraju biti eksportovana - kako bi ih "html/template" paket pročitao prilikom renderovanja
type snippetCreateForm struct {
	Title       string
	Content     string
	Expires     int
	FieldErrors map[string]string
}

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

	// "r.PostForm.Get()" metoda uvijek vraća podatke iz forme u vidu String-a
	// međutim, u našem konkretnom slučaju - vrijednost za "expired" mora biti "integer" i zbog toga vršimo provjeru
	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// instanca "snippetCreateForm" struct-a
	// ona će sadržati vrijednosti iz forme i praznu mapu - koja treba biti popunjavana greškama prilikom validacije
	form := snippetCreateForm{
		Title:       r.PostForm.Get("title"),
		Content:     r.PostForm.Get("content"),
		Expires:     expires,
		FieldErrors: map[string]string{},
	}

	// ukoliko trebamo da vadimo više vrijednosti odjednom (checkbox i slično) preko "r.PostForm()", onda to možemo da odradimo preko petlje
	/*
		for i, item := range r.PostForm["items"] {
			fmt.Fprintf(w, "%d: Item %s\n", i, item)
		}
	*/

	// odličan blog post koji sadrži generalne metode za validaciju:
	// https://www.alexedwards.net/blog/validation-snippets-for-go#email-validation
	// provjera:
	if strings.TrimSpace(form.Title) == "" {
		form.FieldErrors["title"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Title) > 100 {
		form.FieldErrors["title"] = "This field cannot be more than 100 characters long"
	}

	if strings.TrimSpace(form.Content) == "" {
		form.FieldErrors["content"] = "This field cannot be blank"
	}

	if form.Expires != 1 && form.Expires != 7 && form.Expires != 365 {
		form.FieldErrors["expires"] = "This field must equal 1, 7 or 365"
	}

	// ukoliko neke od grešaka postoje, onda treba nanovo prikazati "create.tmpl" templejt
	// dinamički podaci će biti proslijeđeni u "Form" polje
	// takođe, treba poslati 402 HTTP status kod - koji pokazuje da je došlo do greške prilikom validacije
	if len(form.FieldErrors) > 0 {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl", data)
		return

	}

	// prosljeđivanje podataka ka bazi
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// "redirect" putanja mora da se ažurira, kako bi se koristio novi, čistiji URL format
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
