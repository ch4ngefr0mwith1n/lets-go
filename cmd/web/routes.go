package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

// request-ovi u prvobitnoj verziji će biti proslijeđeni na sledeći šablon "secureHeaders → servemux → application handler"
// kada se vrati zadnji "handler" u lancu, onda se kontrola vraća nazad - u kontra smjeru
// secureHeaders → servemux → application handler → servemux → secureHeaders
/*
	func myMiddleware(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Any code here will execute on the way down the chain.
			next.ServeHTTP(w, r)
			// Any code here will execute on the way back up the chain.
    	})
	}
*/

// povratna vrijednost "routes()" metode sada treba da bude "http.Handler"
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// kreira se "handler" funkcija - koja omotava "notFound" helper funkciju
	// nakon toga, postavljamo je kao "custom handler" za "404 Not Found" odgovore
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// kreiranje "file" servera - za fajlove iz "ui/static" foldera
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// sada je "file" server uvezan sa handler-om, koji pokriva sve putanje koje počinju sa "/static/"
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	router.HandlerFunc(http.MethodGet, "/", app.home)
	// obje naredne putanje su fiksne putanje
	// ne završavaju se sa "/"
	router.HandlerFunc(http.MethodGet, "/snippet/view/:id", app.snippetView)
	router.HandlerFunc(http.MethodGet, "/snippet/create", app.snippetCreate)
	router.HandlerFunc(http.MethodPost, "/snippet/create", app.snippetCreatePost)

	// izvršavanje svih "middleware"-a dok se ne dođe do "router"-a
	// stara verzija - app.recoverPanic(app.loqRequest(secureHeaders(mux)))

	// korišćenje "justinas/alice" paketa za nadovezivanje "middleware"-a
	// ima fleksibilnije slučajeve korišćenja
	/*
		myChain := alice.New(myMiddlewareOne, myMiddlewareTwo)
		myOtherChain := myChain.Append(myMiddleware3)
		return myOtherChain.Then(myHandler)
	*/
	standard := alice.New(app.recoverPanic, app.loqRequest, secureHeaders)
	// metodu "then" dodajemo na kraju i njen parametar je router
	return standard.Then(router)
}

// ukoliko se odradi "return" prije narednog poziva "next.ServeHTTP()" - onda se prekida lanac izvršavanja
// kontrola se vraća uzvodno
/*
	func myMiddleware(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If the user isn't authorized, send a 403 Forbidden status and
			// return to stop executing the chain.
			if !isAuthorized(r) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Otherwise, call the next handler in the chain.
			next.ServeHTTP(w, r)
		})
	}
*/
