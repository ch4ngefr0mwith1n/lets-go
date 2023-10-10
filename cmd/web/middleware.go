package main

import (
	"fmt"
	"github.com/justinas/nosurf"
	"net/http"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

// ovo će biti metoda nad "application" struct-om
// razlog - zato što tako ima pristup njegovim zavisnostima, uključujući i "structured logger"
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			method = r.Method
			proto  = r.Proto
			uri    = r.URL.RequestURI()
		)

		// ovdje je napravljena veza sa "structured logger"-om:
		app.logger.Info("received request:", "ip", ip, "method", method, "proto", proto, "uri", uri)
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// kreira se funkcija sa odloženim izvršavanjem
		// ona će uvijek da se izvršava u "panic" slučaju, kako Golang bude "odmotavao" svoj stack
		defer func() {
			// pozivom "recover" funkcije provjeravamo da li se desio "panic" ili ne
			if err := recover(); err != nil {
				// "Connection: close" header se stavlja na odgovor
				// kada se ovaj header podesi, Golang HTTP server će automatski da zatvori trenutnu konekciju
				w.Header().Set("Connection", "close")
				// vraćanje "500 Internal Server" odgovora
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// želimo da izbjegnemo situaciju da korisnik koji nije ulogovan može da kreira "snippet"
// a generalno, trebanam neki "middleware" koji ograničava pristup rutama
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ukoliko korisnik nije ulogovan, trebamo ga preusmjeravati na "login" stranicu
		// nakon toga, treba izaći iz "middleware" lanca
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		}

		// BITNO:
		// stranice koje zahtjevaju da korisnik bude ulogovan NE SMIJU biti sačuvane unutar "cache"-a u browser-u
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// kako bismo izbjegli CSRF, koristićemo "nosurf" paket
// koristi se "double-submit cookie" pristup
// prvo se generiše "random CSRF token" i šalje korisniku unutar "CSRF cookie"-a
// nakon toga, ovaj "CSRF token" se dodaje u svako HTML polje koje je ranjivo na CSRF
// pri slanju forme, koristi se neki "middleware" - koji provjerava da li se skrivena vrijednost u HTML polju i vrijednost "cookie"-a poklapaju

// ova "middleware" funkcija koristi prilagođeni CSRF cookie, unutar kog su postavljene vrijednosti na tri atributa
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		// ovom "cookie"-ju se može pristupiti samo preko HTTP "request"-ova (ne može preko Javascript-a)
		HttpOnly: true,
		// putanja na kojoj je "cookie" validan
		// pošto je ona podešena na "root" - to znači da će se slati sa svim "request"-ovima ka sajtu
		Path: "/",
		// "cookie" treba da se šalje samo preko "HTTPS"-a
		Secure: true,
	})

	return csrfHandler
}
