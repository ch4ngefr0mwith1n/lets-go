package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

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
	// dodavanje "users" polja
	users *models.UserModel
	// dodavanje templateCache polja
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
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

	// nova instanca "decoder"-a:
	formDecoder := form.NewDecoder()

	// inicijalizacija novog "session manager"-a
	sessionManager := scs.New()
	// nakon toga, podešavamo MySQL bazu da služi kao "session store"
	sessionManager.Store = mysqlstore.New(db)
	// i zadnje, podešavamo sesije tako da traju 12 sati nakon vremena kreiranja
	sessionManager.Lifetime = 12 * time.Hour
	// nakon što smo ubacili TLS sertifikat, sada moramo da postavimo "Secure" atribut na "session cookies"
	// to znači da će "cookie" biti poslat iz korisničkog browsera samo prilikom korišćenja HTTPS konekcije
	sessionManager.Cookie.Secure = true

	app := &application{
		logger: logger,
		// inicijalizovanje "models.SnippetModel" instance, koja sadrži "connection pool"
		// nakon toga, dodajemo je u zavisnosti aplikacije
		snippets: &models.SnippetModel{DB: db},
		// isti pristup i sa "users"
		users: &models.UserModel{DB: db},
		// inicijalizovanje "template cache"-a
		templateCache: templateCache,
		// dodavanje instance "decoder"-a u "application" zavisnosti:
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// modifikacija za "TLS elliptic curves" - koje se koriste prilikom TLS "handshake"-a
	// na ovaj način smanjujemo opterećenje servera
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:    *addr,
		Handler: app.routes(),
		// ne možemo direktno da koristimo naš već postojeći "structured logger" handler
		// moramo da ga prebacimo u "*log.Logger", koji upisuje logove na određenom, fiksnom nivou
		ErrorLog:  slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,
		// GoLang omogućava "keep-alive" na svim prihvaćenim konekcijama
		// konekcija između klijenta i servera će biti otvorena određenno vrijeme
		// klijent može koristi istu konekciju za više zahtjeva, bez potrebe za ponavljanjem TLS "handshake"-ova
		// dužina "keep-alive" konekcije zavisi od OS-a kog koristimo, ne možemo da je povećamo
		// međutim, možemo da je smanjimo preko "idleTimeout"
		IdleTimeout: time.Minute,
		// ukoliko se "request headers" ili "request body" učitavaju i obrađuju duže od 5 sekundi nakon prihvatanja zahtjeva, Golang će zatvoriti konekciju
		// ovo je "hard closure" sa serverske strane, pa korisnik neće dobiti nikakav HTTP(S) odgovor
		// na ovaj način se smanjuje rizik od "slow client" napada, poput Slowloris-a (veza bi bila otvorena non-stop, zbog slanja nepotpunih HTTP(S) zahtjeva)
		// BITNO:
		// ukoliko postavimo vrijednost za "ReadTimeout", a eksplicitno ne postavimo "IdleTimeout" - onda će njihove vrijednosti biti jednake
		ReadTimeout: 5 * time.Second,
		// "WriteTimeout" uvijek mora da ima veću vrijednost od "ReadTimeout"-a (učitavanje "request"-a, pa ispisivanje "response"-a)
		// to je zato što se u HTTPS vrijeme računa od prihvatanja zahtjeva (kod HTTP-a se računa od učitavanja "request header"-a)
		WriteTimeout: 10 * time.Second,
		// vrijeme potrebno za učitavanje "header"-a
		// učitavanje "request body"-ja može da traje duže od ovoga, bez zatvaranja konekcije
		ReadHeaderTimeout: 3 * time.Second,
	}

	// logger obavještava da će server biti pokrenut
	logger.Info(fmt.Sprintf("Starting server on port %s", *addr))
	// koristićemo novu metodu za pokretanje HTTPS servera
	// moramo da proslijedimo "public" i "private" key kao parametre
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	// u slučaju da se desi neka greška - aplikacija će biti izgašena
	logger.Error(err.Error())
	os.Exit(1)
}
