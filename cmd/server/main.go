package main

import (
	"log"
	"net/http"

	"github.com/tfharrelson/scicomp-bench/app/app"
)

func main() {
	mux := app.CreateApp()

	port := ":8080"
	log.Printf("Server starting on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
