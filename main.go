package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	initializeBlogPosts()
	router := mux.NewRouter()
	initializeRoutes(router)
	log.Fatal(http.ListenAndServe(":8080", router))
}
