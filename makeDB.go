package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type userBaseDetails struct {
    FirstName string
    SecondName string
    Username string
    Pwd string
}

var users = []userBaseDetails{
    {"Robert", "Gascoyne-Cecil", "salisbury", "1895"},
    {"Arthur", "Balfour", "abalfour", "1902"},
    {"Henry", "Campbell-Bannerman", "hcb", "1905"},
    {"Herbert", "Asquith", "hasquith", "1908"},
    {"David", "Lloyd George", "dlg", "1916"},
}

func main() {
	os.Remove("./jerver.db")

	db, err := sql.Open("sqlite3", "./jerver.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
    for _, user := range(users) {
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
}
