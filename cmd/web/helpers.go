package main

import (
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
