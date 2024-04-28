package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"graphviz-renderer-web/server"
)

func main() {
	log.Print("Chart server starting up")

	http.HandleFunc("/chart", server.RenderGV)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
