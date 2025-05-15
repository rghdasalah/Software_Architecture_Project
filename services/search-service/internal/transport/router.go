package transport

import (
	//"net/http"

	"github.com/gorilla/mux"
)

func SetupRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	// Register your search endpoint
	router.HandleFunc("/search", handler.SearchRidesHandler).Methods("POST")

	return router
}
