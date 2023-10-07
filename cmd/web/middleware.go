package main

import (
	"fmt"
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
