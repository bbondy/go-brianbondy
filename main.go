package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	initializeBlogPosts()
	router := mux.NewRouter()
	initializeRoutes(router)
	http.ListenAndServe(":8080", router)
}
