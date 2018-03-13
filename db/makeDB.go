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
	Title string
}

var threads = []threadBaseDetails{
	{"Who's the best PM?"},
	{"Favourite Commons memory?"},
}

type messageBaseDetails struct {
	ThreadIndex uint
	AuthorIndex uint
	Content     string
}

var messages = []messageBaseDetails{
	{0, 3, "Asquith"},
	{0, 4, "Lloyd George"},
	{1, 3, "That day we declared war"},
	{1, 4, "When I forced Asquith out"},
}

func main() {
	os.Remove("../jerver.db")

	db, err := sql.Open("sqlite3", "../jerver.db")
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

	userUuids := []uuid.UUID{}
	for i, user := range users {
		userUuid, _ := uuid.NewV4()
		userUuids = append(userUuids, userUuid)
		hpwd, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		_, err = stmt.Exec(userUuids[i].Bytes(), user.FirstName,
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

	threadUuids := []uuid.UUID{}
	for i, thread := range threads {
		threadUuid, _ := uuid.NewV4()
		threadUuids = append(threadUuids, threadUuid)
		_, err = stmt.Exec(threadUuids[i].Bytes(), thread.Title)
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()

	// CREATE MESSAGES TABLE
	sqlStmt = `
    CREATE TABLE messages (
        Uuid blob NOT NULL PRIMARY KEY,
        ThreadId blob NOT NULL,
        AuthorId blob NOT NULL,
        Content string,
        FOREIGN KEY(ThreadId) REFERENCES threads(Uuid),
        FOREIGN KEY(AuthorId) REFERENCES users(Uuid))
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	// POPULATE MESSAGES TABLE
	tx, err = db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err = tx.Prepare(`
    INSERT INTO messages(
        Uuid,
        ThreadId,
        AuthorId,
        Content)
    VALUES (?, ?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, message := range messages {
		messageUuid, _ := uuid.NewV4()
		_, err = stmt.Exec(
			messageUuid.Bytes(),
			threadUuids[message.ThreadIndex].Bytes(),
			userUuids[message.AuthorIndex].Bytes(),
			message.Content)

		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
}
