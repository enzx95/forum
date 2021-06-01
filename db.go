package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Specie struct {
	Id     int
	TypeId int
	Name   string
}

type Type struct {
	Id   int
	Name string
}

// func initDB(database string) *sql.DB {
// 	db, err := sql.Open("sqlite3", database)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	sqltStmt := `
// 				PRAGMA foreign_keys = ON;
// 				CREATE TABLE IF NOT EXISTS types(
// 					id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT ,
// 					name TEXT NOT NULL
// 				);

// 				CREATE TABLE IF NOT EXISTS species(
// 					id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT ,
// 					type_id INTEGER NOT NULL  ,
// 					name TEXT NOT NULL,
// 					FOREIGN KEY (type_id) REFERENCES types(id)
// 				);
// 				`

// 	_, err = db.Exec(sqltStmt)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return db
// }

// func insertType(db *sql.DB, name string) (int64, error) {
// 	result, _ := db.Exec(`INSERT INTO types (name) VALUES (?)`, name)
// 	return result.LastInsertId()
// }

func insertSpecies(db *sql.DB, name string, type_id int) (int64, error) {
	result, _ := db.Exec(`INSERT INTO species (type_id, name) VALUES (?, ?)`, type_id, name)
	return result.LastInsertId()
}

// func selectAllFromTables(db *sql.DB, table string) *sql.Rows {
// 	query := "SELECT * FROM " + table
// 	result, _ := db.Query(query)
// 	return result

// }

func selectTypesById(db *sql.DB, id int) Type {
	var t Type
	db.QueryRow(`SELECT * FROM types WHERE id = ?`, id).Scan(&t.Id, &t.Name)
	return t
}

func displayTypeRows(rows *sql.Rows) {
	for rows.Next() {
		var u Type
		err := rows.Scan(&u.Id, &u.Name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(u)
	}

}

func displaySpecieRows(rows *sql.Rows) {
	for rows.Next() {
		var p Specie
		err := rows.Scan(&p.Id, &p.TypeId, &p.Name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(p)
	}

}

// func main() {
// 	db := initDB("test.db")
// 	defer db.Close()
// 	insertType(db, "Zeub")
// 	insertType(db, "Grinch")
// 	insertSpecies(db, "Kriba", 1)
// 	insertSpecies(db, "Doumams", 2)
// 	rowsTypes := selectAllFromTables(db, "types")
// 	displayTypeRows(rowsTypes)
// 	speciesTypes := selectAllFromTables(db, "species")
// 	displaySpecieRows(speciesTypes)
// 	//fmt.Println(db)
// }
