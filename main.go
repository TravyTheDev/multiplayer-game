package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	router := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
	})

	handler := c.Handler(router)
	hub := NewHub()
	wsHandler := NewWSHandler(hub)
	wsHandler.registerRoutes(router)
	go hub.Run()

	fmt.Println("Listening on :8000")
	http.ListenAndServe(":8000", handler)
}
