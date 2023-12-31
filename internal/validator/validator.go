package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors map[string]string
	// u ovom polju ćemo čuvati greške prilikom validacije koje nisu vezane za polja unutar forme
	NonFieldErrors []string
}

// provjera da "FieldErrors" mapa ne sadrži nijedan unos:
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// metoda za dodavanje "error" poruke u "FieldErrors" mapu
func (v *Validator) AddFieldErrorKey(key string, message string) {
	// ukoliko mapa već nije inicijalizovana, inicijalizovaćemo je:
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	// provjeravamo da li navedeni "key:value" unos već postoji:
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// ova metoda dodaje "error message" u FieldErrors mapu samo ako validacija ne prolazi
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldErrorKey(key, message)
	}
}

// "helper" metoda za dodavanje "error" poruka u "NonFieldsErrors" slice
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// ova funkcija vraća "true" ako vrijednost pripada listi dozvoljenih vrijednosti
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
