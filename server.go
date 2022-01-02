package main

import (
	"fmt"
	"github.com/ashishkumar68/aws-ops/awsClient"
	"github.com/ashishkumar68/aws-ops/controller"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	// start resnap message listener
	snapshotService := awsClient.NewRDSSnapshotService()
	go snapshotService.RunResnapMessageListener()

	port := os.Getenv("HTTP_PORT")
	r := mux.NewRouter()
	r.HandleFunc("/resnap-database", controller.ResnapRDSByName).Methods("POST")

	http.Handle("/", r)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		log.Println("Could not start the HTTP server.")
		log.Println(err)
		return
	}

	log.Println("running server on port", port)
}