package main

import (
	"log"
	"net/http"
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
	w.Write([]byte("Display a specific snippet"))
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
