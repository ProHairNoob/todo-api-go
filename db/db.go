package db

import (
	"database/sql"
	"errors"
	"fmt"
	//"log"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
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

var ErrUserExists = errors.New("username or email already exists")

func InsertUser(dbConn *sql.DB, user string, email string, hash string) (int64, error) {
	result, err := dbConn.Exec("INSERT INTO users (username,email,password_hash) VALUES (?,?,?)",
		user, email, hash,
	)
	if err != nil {
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return 0, ErrUserExists
		}
		return 0, err
	}

	UserID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return UserID, err
}

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./api.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		fmt.Println(err)
		return db, err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return db, err
	}
	db.SetMaxOpenConns(1)
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
