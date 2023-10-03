package models

import (
	"database/sql"
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

type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Get(id int) (Snippet, error) {
	var snippet Snippet
	return snippet, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	return nil, nil
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	return 0, nil
}
