package main

import "net/http"

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
	mux := http.NewServeMux()

	// kreiranje "file" servera - za fajlove iz "ui/static" foldera
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// sada je "file" server uvezan sa handler-om, koji pokriva sve putanje koje počinju sa "/static/"
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", app.home)
	// obje naredne putanje su fiksne putanje
	// ne završavaju se sa "/"
	mux.HandleFunc("/snippet/view", app.snippetView)
	mux.HandleFunc("/snippet/create", app.snippetCreate)

	// "logRequest" middleware treba prvi da se izvrši, a nakon njega će ići "secureHeaders", "router" pa "handler"-i
	return app.loqRequest(secureHeaders(mux))
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
