package main

import (
	chi "github.com/go-chi/chi/v5"
)

func (app *application) setupRoutes() chi.Router {
	
	router := chi.NewRouter()
	router.Get("/healthcheck", app.healthcheck)
	router.Post("/validate", app.validate)
	router.Get("/healthz", app.healthcheck)
	return router
}
