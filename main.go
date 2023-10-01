package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func home(w http.ResponseWriter, r *http.Request) {
	// "/" putanja ne treba da "hvata" sve zahtjeve
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Write([]byte("Hello from snippetbox!"))
}

func snippetView(w http.ResponseWriter, r *http.Request) {
	// vadi se vrijednost "id" parametra iz URL-a
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
	}

	// ova funkcija prima "io.writer" interfejs
	// "http.ResponseWriter" zadovoljava uslove ovog interfejsa
	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

func snippetCreate(w http.ResponseWriter, r *http.Request) {
	// po default-u, "status code" koji se šalje je "200 OK"
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		//http.Error(w, "Method not allowed", 405)

		//w.WriteHeader(405)
		//w.Write([]byte("Method not allowed!"))
		return
	}

	w.Write([]byte("Creating a new snippet..."))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", home)
	// obje naredne putanje su fiksne putanje
	// ne završavaju se sa "/"
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	log.Println("Starting server on port :4000 ")

	// pokretanje servera
	// u parametre idu adresa i router
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
