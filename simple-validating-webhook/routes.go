package main

import (
	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {

	router := mux.NewRouter()
	router.HandleFunc("/healthcheck", app.healthcheck).Methods("GET")
	router.HandleFunc("/validate", app.validate).Methods("POST")

	return router

}
