package main

import (
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from snippetbox!"))
}

func main() {
	mux := http.NewServeMux()
	// trenutno, svaki request će ići na ovaj endpoint
	// čak i request-ovi poput "http://localhost:4000/foo"
	mux.HandleFunc("/", home)

	log.Println("Starting server on port :4000 ")

	// pokretanje servera
	// u parametre idu adresa i router
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
