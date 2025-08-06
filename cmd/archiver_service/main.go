package main

import (
	"net/http"
	"log"
	"time"
	"fmt"
	"html"
)

func main() {
	// TODO: Ref, move to config
	s := &http.Server{
		Addr: ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("\033[32mINFO\033[32m %s %s", r.Method, r.URL.Path) // ip
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Fatal(s.ListenAndServe())
}