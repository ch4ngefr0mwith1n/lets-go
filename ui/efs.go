package ui

import "embed"

// BITNO:
// "embed" biblioteka nam služi da "ugradimo" eksterne fajlove u sam Golang program
// u našem slučaju, poslužiće nam za "ugrađivanje" fajlova iz "ui" direktorijuma - statički CSS, JavaScript, slike...
// nakon toga, svi ti fajlovi će biti ukomponovani u HTML templejte, koje inače koristimo

// "go:embed "static" predstavlja poseban komentar
// odmah nakon kompajliranja programa, fajlovi iz "ui/static" foldera će biti sačuvani u "embedded" fajl sistemu
// globalna promjenjiva "Files" referencira taj fajl sistem

//go:embed "html" "static"
var Files embed.FS
