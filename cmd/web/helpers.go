package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// "serverError" piše "log entry" na "Error" nivou
// nakon toga, šalje "500 Internal Server Error" odgovor ka korisniku
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		// koristi se "debug.Stack()" da se dobije "stack trace"
		trace = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// "clientError" šalje određeni status kod i odgovarajući opis ka korisniku
// koristiće se u slučajevima kada postoji problem u "request"-u koji je korisnik poslao
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// "Not found" helper
// varijacija "clientError"-a, koja šalje "404 Not Found" odgovor ka klijentu
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	// Retrieve the appropriate template set from the cache based on the page
	// name (like 'home.tmpl'). If no entry exists in the cache with the
	// provided name, then create a new error and call the serverError() helper
	// method that we made earlier and return.
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// Write out the provided HTTP status code ('200 OK', '400 Bad Request' etc).
	w.WriteHeader(status)

	// Execute the template set and write the response body. Again, if there
	// is any error we call the the serverError() helper.
	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}
