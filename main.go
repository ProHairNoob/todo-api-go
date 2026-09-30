package main

import (
	"log"
	"net/http"
	"time"

	"todo-api/api"
	"todo-api/db"
)

func main() {
	database, err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	defer database.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register ", api.RegisterHandler(database))
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
