package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// kreiranje "file" servera - za fajlove iz "ui/static" foldera
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// sada je "file" server uvezan sa handler-om, koji pokriva sve putanje koje počinju sa "/static/"
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", home)
	// obje naredne putanje su fiksne putanje
	// ne završavaju se sa "/"
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	log.Println("Starting server on port :4000 ")

	// pokretanje servera
	// u parametre idu adresa i router
	err := http.ListenAndServe("127.0.0.1:4000", mux)
	log.Fatal(err)
}
