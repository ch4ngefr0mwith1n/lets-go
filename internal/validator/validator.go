package validator

import (
	"slices"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors map[string]string
}

// provjera da "FieldErrors" mapa ne sadrži nijedan unos:
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
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