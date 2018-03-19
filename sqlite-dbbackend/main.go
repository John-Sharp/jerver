package dbbackend

import (
	"database/sql"
	"github.com/john-sharp/jerver/entities"
	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"log"
)

var db *sql.DB
var getMessageFromUuidStmt *sql.Stmt

func init() {
	var err error

	db, err = sql.Open("sqlite3", "./jerver.db")
	if err != nil {
		log.Fatal(err)
	}

	messagePrepareStmts()
}

func messagePrepareStmts() {
	var err error
	getMessageFromUuidStmt, err = db.Prepare(`
	SELECT
	     Uuid,
	     ThreadId,
	     AuthorId,
	     Content
	FROM messages
	WHERE Uuid = ?`)

	if err != nil {
		log.Fatal(err)
	}
}

func GetMessageFromUuid(targetUuid uuid.UUID) (*entities.Message, error) {
	var m entities.Message
	err := getMessageFromUuidStmt.QueryRow(targetUuid.Bytes()).Scan(&m.Id, &m.ThreadId, &m.AuthorId, &m.Content)

	if err != nil {
		return nil, err
	}
	return &m, nil
}
