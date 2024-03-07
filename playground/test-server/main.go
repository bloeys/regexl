package main

import (
	"log"
	"net/http"
)

func main() {

	fs := http.FileServer(http.Dir("../"))
	http.Handle("/", fs)

	log.Print("Listening on localhost:3434...")
	err := http.ListenAndServe(":3434", nil)
	if err != nil {
		log.Fatal(err)
	}
}
