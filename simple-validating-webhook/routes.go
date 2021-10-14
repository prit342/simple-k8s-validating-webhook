package main

//// newRouter - returns a new mux
//func newRouter() *mux.Router {
//	return mux.NewRouter()
//}
//
//
//// routeTable - various paths in the route table along with http method
//type routeTable struct{
//	path string
//	handler http.HandlerFunc
//	httpMethod string
//}
//
//
//// setUpRoutes - sets up routes and binds them to the mux
//func (app *application) setUpRoutes(r *mux.Router, routes []routeTable)  error {
//
//	if r == nil {
//		return fmt.Errorf("http mux cannot be nil")
//	}
//
//	if len(routes) < 1 {
//		return fmt.Errorf("size of route table can not be zero")
//	}
//
//
//	for _, route := range routes {
//		r.HandleFunc(route.path, route.handler).Methods(route.httpMethod)
//	}
//
//	return nil
//}

import (
	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {

	router := mux.NewRouter()
	router.HandleFunc("/healthcheck", app.healthcheck).Methods("GET")
	router.HandleFunc("/validate", app.validate).Methods("POST")
	router.HandleFunc("/healthz", app.healthcheck).Methods("GET")
	return router

}
