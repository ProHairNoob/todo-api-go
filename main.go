package main

import (
	//"fmt"
	"io"
	"log"
	"net/http"

	"todo-api/api"
	"todo-api/db"
)

func main() {
	mux := http.ServeMux

	database, err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	http.HandleFunc("POST /register", api.RegisterHandler(database))
	http.HandleFunc("GET /{$}", h1)
	http.HandleFunc("/foo", h2)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
