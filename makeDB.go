package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

type userBaseDetails struct {
	FirstName  string
	SecondName string
	Username   string
	Pwd        string
}

var users = []userBaseDetails{
	{"Robert", "Gascoyne-Cecil", "salisbury", "1895"},
	{"Arthur", "Balfour", "abalfour", "1902"},
	{"Henry", "Campbell-Bannerman", "hcb", "1905"},
	{"Herbert", "Asquith", "hasquith", "1908"},
	{"David", "Lloyd George", "dlg", "1916"},
}

type threadBaseDetails struct {
    Title  string
}

var threads = []threadBaseDetails{
    {"Who's the best PM?"},
    {"Favourite Commons memory?"},
}

func main() {
	os.Remove("./jerver.db")

	db, err := sql.Open("sqlite3", "./jerver.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

    // CREATE USERS TABLE
	sqlStmt := `
	CREATE TABLE users (
        Uuid blob NOT NULL PRIMARY KEY, 
        FirstName text,
        SecondName text,
        Username text,
        HashedPwd blob);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

    // POPULATE USERS TABLE
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare(`
    INSERT INTO users(
        Uuid,
        FirstName,
        SecondName,
        Username,
        HashedPwd)
    VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for _, user := range users {
		hpwd, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		_, err = stmt.Exec(uuid.NewV4().Bytes(), user.FirstName,
			user.SecondName, user.Username, hpwd)
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
    
    // CREATE THREADS TABLE
	sqlStmt = `
	CREATE TABLE threads (
        Uuid blob NOT NULL PRIMARY KEY, 
        Title text);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

    // POPULATE THREADS TABLE
	tx, err = db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err = tx.Prepare(`
    INSERT INTO threads(
        Uuid,
        Title)
    VALUES (?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for _, thread := range threads {
		_, err = stmt.Exec(uuid.NewV4().Bytes(), thread.Title)
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
}
