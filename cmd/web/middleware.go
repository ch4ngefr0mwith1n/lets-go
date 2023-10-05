package main

import "net/http"

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

// ovoće biti metoda nad "application" struct-om
// razlog - zato što tako ima pristup njegovim zavisnostima, uključujući i "structured logger"
func (app *application) loqRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			method = r.Method
			proto  = r.Proto
			uri    = r.URL.RequestURI()
		)

		app.logger.Info("received request:", "ip", ip, "method", method, "proto", proto, "uri", uri)
		next.ServeHTTP(w, r)
	})
}
