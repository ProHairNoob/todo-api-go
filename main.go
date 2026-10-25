package main

import (
	"log"
	"net/http"
	"time"

	"todo-api/api"
	"todo-api/db"
)

func main() {
	api.LoadEnv()
	database, err := db.InitDB("./api.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", api.RegisterHandler(database))
	mux.HandleFunc("POST /login", api.LoginHandler(database))
	mux.HandleFunc("POST /todo", api.AuthMiddleware(api.TodoHandler(database)))
	mux.HandleFunc("DELETE /todo/{task_id}", api.AuthMiddleware(api.DeleteTodoHandler(database)))
	mux.HandleFunc("PUT /todo/{task_id}", api.AuthMiddleware(api.UpdateTodoHandler(database)))
	mux.HandleFunc("GET /todo", api.AuthMiddleware(api.GetTodoHandler(database)))
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
