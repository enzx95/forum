package database

import "database/sql"

var DB *sql.DB

func InitDB() *sql.DB {
	db, _ := sql.Open("sqlite3", "database.db")
	db.Exec(`create table if not exists users (username text NOT NULL, email text NOT NULL, password text NOT NULL)`)
	db.Exec(`create table if not exists likes (author text NOT NULL, numpost text NOT NULL, date text NOT NULL)`)
	db.Exec(`create table if not exists replylikes (author text NOT NULL, numpost text NOT NULL, date text NOT NULL)`)
	db.Exec(`create table if not exists dislikes (author text NOT NULL, numpost text NOT NULL, date text NOT NULL)`)
	db.Exec(`create table if not exists replydislikes (author text NOT NULL, numpost text NOT NULL, date text NOT NULL)`)
	db.Exec(`create table if not exists posts (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, author text NOT NULL, content text NOT NULL, title text NOT NULL, created text NOT NULL, categories text NOT NULL)`)
	db.Exec(`create table if not exists replies (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT , numpost integer NOT NULL, author text NOT NULL, content text NOT NULL,  created text NOT NULL)`)
	return db
}

func SelectAllFromTables(db *sql.DB, table string) *sql.Rows {
	query := "SELECT * FROM " + table
	result, _ := db.Query(query)
	return result

}
