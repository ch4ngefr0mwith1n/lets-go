package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	// go get github.com/go-sql-driver/mysql
	// potreban nam je ovaj import radi "init" funkcije sadržane unutar njega
	// "_" je alias, moramo da ga koristimo jer ovaj paket nigdje eksplicitno ne koristimo
	_ "github.com/go-sql-driver/mysql"

	"snippetbox.lazarmrkic.com/internal/models"
)

type application struct {
	logger *slog.Logger
	// dodavanje "snippets" polja u "application" struct
	// to će omogućiti da "SnippetModel" objekat bude dostupan kontrolerima
	snippets *models.SnippetModel
	// dodavanje templateCache polja
	templateCache map[string]*template.Template
}

// funkcija za inicijalizovanje "connection pool"-a
// njoj prosljeđujemo DSN string
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// "Ping" kreira testnu konekciju i vrši se provjera grešaka pri prvom testu
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	// kreiranje "command line flag"-a
	addr := flag.String("addr", "127.0.0.1:4000", "HTTP network address")
	// definisanje novog "command line flag"-a, za MySQL DSN (data source name) String
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	// parsiranje flag-a
	flag.Parse()

	// kreiranje novog "structured" logger-a
	// ukoliko želimo da se log čuva u nekom fajlu, onda pokrećemo aplikaciju preko "go run ./cmd/web >>/tmp/web.log"
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// inicijalizovanje "connection pool"-a
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// "connection pool" treba da se zatvori prije izlaska iz "main()" funkcije
	defer db.Close()

	// inicijalizovanje novog "template cache"-a:
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		logger: logger,
		// inicijalizovanje "models.SnippetModel" instance, koja sadrži "connection pool"
		// nakon toga, dodajemo je u zavisnosti aplikacije
		snippets: &models.SnippetModel{DB: db},
		// inicijalizovanje "template cache"-a
		templateCache: templateCache,
	}

	// logger obavještava da će server biti pokrenut
	logger.Info(fmt.Sprintf("Starting server on port %s", *addr))
	// u parametre idu adresa iz "flag"-a i router
	err = http.ListenAndServe(*addr, app.routes())
	// u slučaju da se desi neka greška - aplikacija će biti izgašena
	logger.Error(err.Error())
	os.Exit(1)
}
