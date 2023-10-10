package main

import (
	"snippetbox.lazarmrkic.com/internal/assert"
	"testing"
	"time"
)

// BITNO:
// svaka test metoda u Golang-u ima potpis "func(*testing.T)"
// kako bi test bio VALIDAN, naziv funkcije mora da počinje sa "Test"
// poziv funkcije "t.Errorf()" neće zaustaviti izvršavanje ostatka testova
//
// komanda za pokretanje testova - "go test ./cmd/web/"
// dodajemo "-v" ukoliko zelimo da vidimo koji tačno testovi se izvršavaju - "go test -v ./cmd/web"
//
// BITNO:
// "subtest"-ove ćemo praviti na sledeći način:
//
//	func TestExample(t *testing.T) {
//	    t.Run("Example sub-test 1", func(t *testing.T) {
//	        // Do a test.
//	    })
//
//	    t.Run("Example sub-test 2", func(t *testing.T) {
//	        // Do another test.
//	    })
//
//	    t.Run("Example sub-test 3", func(t *testing.T) {
//	        // And another...
//	    })
//	}
func TestHumanDate1(t *testing.T) {
	tm := time.Date(2023, 3, 17, 10, 15, 0, 0, time.UTC)
	hd := humanDate(tm)

	// ukoliko dobijemo rezultat koji ne očekujemo, onda koristimo "t.Errorf()" funkciju
	// ona pokazuje da je test propao i pravi log u kom su očekivana i stvarna vrijednost
	if hd != "17 Mar 2023 at 10:15" {
		t.Errorf("got %q; want %q", hd, "17 Mar 2023 at 10:15")
	}
}

func TestHumanDate2(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2023, 3, 17, 10, 15, 0, 0, time.UTC),
			want: "17 Mar 2023 at 10:15",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2023, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17 Mar 2023 at 09:15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)

			// sada ćemo koristiti "assert.Equal" za testiranje
			assert.Equal(t, hd, tt.want)
		})
	}
}
