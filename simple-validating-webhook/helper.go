package main

import (
	"fmt"
	"net/http"
)

// writeErrorMessage - writes error message to stderr and the http stream
func (app *application) writeErrorMessage(w http.ResponseWriter, msg string, code int) {
	
	w.Header().Set("Content-Type", "application/json")
	app.errorLog.Println(msg)
	msg = fmt.Sprintf(`{"error": "%v"}`, msg)
	//http.Error(w, msg, http.StatusInternalServerError)
	http.Error(w, msg, code)
	
}
