package models

import (
	"database/sql"
	"errors"
	"time"
)

// "internal" direktorijum će sadržati sav "non-application specific" kod
// odnosno, kod koji potencijalno opet može da se koristi
// "database model" može da se koristi od strane drugih aplikacija u budućnosti

// "models" paket predstavlja "service" / "data access" layer
// u ovom paketu ćemo enkapsulirati kod za rad sa MySQL-om

// polja Snippet "struct"-a će da odgovaraju poljima MySQL tabele
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// deklarisanjem ovog tipa i implementiranjem metoda nad njim - imamo jedan enkapsulirani objekat
// lako možemo da ga inicijalizujemo i nakon toga, da proslijedimo u "handler"-e kao zavisnost
type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Get(id int) (Snippet, error) {

	stmt := `SELECT id, title, content, created, expires FROM snippets WHERE expires > UTC_TIMESTAMP() AND id = ?`
	row := m.DB.QueryRow(stmt, id)

	var s Snippet

	// metodu "row.Scan()" koristimo da kopiramo vrijednosti iz "sql.Row" polja u odgovarajuće polje u "Snippet" struct-u
	// parametri ove metode su "pointer"-i ka mjestima gdje želimo da iskopiramo vrijednosti
	// broj argumenata mora biti isti kao i broj kolona koje naredba vraća
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// ukoliko "query" ne vraća nijedan red, onda će se vratiti "custom" greška - koju smo definisali u "models/errors.go"
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

// skraćena verzija ove metode iznad:
func (m *SnippetModel) GetShorthand(id int) (Snippet, error) {
	var s Snippet

	err := m.DB.QueryRow("SELECT ...", id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	return nil, nil
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires)
    VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// "Exec()" metoda se koristi nad "connection pool"-om, kako bi smo izvršili naredbu
	// ona će vratiti "sql.Result" tip, koji sadrži informacije o izvršavanju naredbe
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// "id" koji generiše baza nakon izvršavanja komande
	// ovaj "id" će generisati "auto increment" kolona
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
