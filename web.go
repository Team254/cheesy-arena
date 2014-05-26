// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Configuration and functions for the event server web interface.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const httpPort = 8080

func ServeWebInterface() {
	http.Handle("/", newHandler())
	log.Printf("Serving HTTP requests on port %d", httpPort)
	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)
}

func newHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/reports/csv/teams", TeamsCsvReportHandler)
	router.HandleFunc("/reports/pdf/teams", TeamsPdfReportHandler)
	return router
}

func handleWebErr(w http.ResponseWriter, err error) {
	http.Error(w, "Internal server error: "+err.Error(), 500)
}
