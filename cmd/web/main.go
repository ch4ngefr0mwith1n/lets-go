package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
)

type application struct {
	logger *slog.Logger
}

func main() {
	// kreiranje "command line flag"-a
	addr := flag.String("addr", "127.0.0.1:4000", "HTTP network address")
	// parsiranje flag-a
	flag.Parse()

	// kreiranje novog "structured" logger-a
	// ukoliko želimo da se log čuva u nekom fajlu, onda pokrećemo aplikaciju preko "go run ./cmd/web >>/tmp/web.log"
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		logger: logger,
	}

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

	// logger obavještava da će server biti pokrenut
	logger.Info("Starting server on port %s", *addr)
	// u parametre idu adresa iz "flag"-a i router
	err := http.ListenAndServe(*addr, mux)
	// u slučaju da se desi neka greška - aplikacija će biti izgašena
	logger.Error(err.Error())
	os.Exit(1)
}
