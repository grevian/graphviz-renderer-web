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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
