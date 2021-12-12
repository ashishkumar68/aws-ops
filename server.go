package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	r := mux.NewRouter()

	http.Handle("/", r)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		fmt.Println("Could not start the HTTP server.")
		fmt.Println(err)
		return
	}

	fmt.Println("running server on port", port)
}