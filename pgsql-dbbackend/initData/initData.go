package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
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
	connStr := "user=jerver dbname=jerver sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
    VALUES ($1, $2, $3, $4, $5)`)
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

		_, err = stmt.Exec(userUuids[i], user.FirstName,
			user.SecondName, user.Username, hpwd)
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()

	// POPULATE THREADS TABLE
	tx, err = db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err = tx.Prepare(`
    INSERT INTO threads(
        Uuid,
        Title)
    VALUES ($1, $2)`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	threadUuids := []uuid.UUID{}
	for i, thread := range threads {
		threadUuid, _ := uuid.NewV4()
		threadUuids = append(threadUuids, threadUuid)
		_, err = stmt.Exec(threadUuids[i], thread.Title)
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()

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
    VALUES ($1, $2, $3, $4)`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, message := range messages {
		messageUuid, _ := uuid.NewV4()
		_, err = stmt.Exec(
			messageUuid,
			threadUuids[message.ThreadIndex],
			userUuids[message.AuthorIndex],
			message.Content)

		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
}
