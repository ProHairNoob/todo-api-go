package db

import (
	"database/sql"
	"fmt"

	//"log"

	_ "modernc.org/sqlite"
)

func createUsersTable(db *sql.DB) error {
	sql := `CREATE TABLE IF NOT EXISTS users(
		user_id INTEGER PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func createTasksTable(db *sql.DB) error {
	sql := `CREATE TABLE IF NOT EXISTS tasks(
		id INTEGER PRIMARY KEY,
		desc TEXT NOT NULL,
		title TEXT NOT NULL,
		status TEXT CHECK(status in ('todo','in-progress','done')) DEFAULT 'todo',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func InsertUser(dbConn *sql.DB, user string, email string, hash string) (int64, error) {
	result, err := dbConn.Exec("INSERT INTO users (username,email,password_hash) VALUES (?,?,?)",
		user, email, hash,
	)
	if err != nil {
		return 0, err
	}
	UserID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return UserID, err
}

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./api.db")
	if err != nil {
		fmt.Println(err)
		return db, err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return db, err
	}
	err = createTasksTable(db)
	if err != nil {
		fmt.Println(err)
		return db, err
	}
	err = createUsersTable(db)
	if err != nil {
		fmt.Println(err)
		return db, err
	}
	return db, err
}
