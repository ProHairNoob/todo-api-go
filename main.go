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
	h1 := func(w http.ResponseWriter, __ *http.Request) {
		io.WriteString(w, "Hello from root!\n")
	}
	h2 := func(w http.ResponseWriter, __ *http.Request) {
		io.WriteString(w, "Hello from foo!\n")
	}
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
