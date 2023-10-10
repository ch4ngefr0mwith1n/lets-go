package assert

import (
	"testing"
)

// prilikom testiranja, većina testova će biti na ovaj šablon:
// if actualValue != expectedValue {
//     t.Errorf("got %v; want %v", actualValue, expectedValue)
// }

// "Equal" je generička funkcija
// koristićemo je bez obzira na tip "actual" i "expected" vrijednosti
func Equal[T comparable](t *testing.T, actual, expected T) {
	// "t.Helper()" funkcija ukazuje da je "Equal()" funkcija zapravo "test helper"
	// odnosno, kada se pozove "t.Errorf()" - prikazaće se "filename" i broj reda u kom se greška nalazi
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}
