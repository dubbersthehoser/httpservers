package main

import (
	"net/http"
	"log"
)


func main() {
	sMux := http.NewServeMux()
	sMux.Handle("/", http.FileServer(http.Dir(".")))
	s := &http.Server{
		Addr: ":8080",
		Handler: sMux,
	}

	log.Fatal(s.ListenAndServe())
}
