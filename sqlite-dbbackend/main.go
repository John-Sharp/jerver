package dbbackend

import (
	"database/sql"
	"github.com/john-sharp/jerver/entities"
	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"log"
)

var db *sql.DB
var getMessageByUuidStmt *sql.Stmt
var createMessageStmt *sql.Stmt
var deleteMessageByUuidStmt *sql.Stmt

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
	getMessageByUuidStmt, err = db.Prepare(`
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

	createMessageStmt, err = db.Prepare(`
    INSERT INTO messages (
        Uuid,
        ThreadId,
        AuthorId,
        Content)
    VALUES (?, ?, ?, ?)`)

	if err != nil {
		log.Fatal(err)
	}

	deleteMessageByUuidStmt, err = db.Prepare(`
    DELETE FROM messages
    WHERE Uuid = ?
    `)

	if err != nil {
		log.Fatal(err)
	}
}

func GetMessageByUuid(targetUuid uuid.UUID) (*entities.Message, error) {
	var m entities.Message
	err := getMessageByUuidStmt.QueryRow(targetUuid.Bytes()).Scan(&m.Id, &m.ThreadId, &m.AuthorId, &m.Content)

	if err != nil {
		return nil, err
	}
	return &m, nil
}

func CreateMessage(m *entities.Message) error {
	_, err := createMessageStmt.Exec(
		m.Id.Bytes(),
		m.ThreadId.Bytes(),
		m.AuthorId.Bytes(),
		m.Content)

	return err
}

func DeleteMessageByUuid(targetUuid uuid.UUID) error {
	_, err := deleteMessageByUuidStmt.Exec(targetUuid.Bytes())
	return err
}

func GetMessageCollection(threadId uuid.UUID, count uint64, page int64, appendToCollection func(entities.Message)) error {
	offset := page * int64(count)

	// TODO put in filtering
	rows, err := db.Query(`
    SELECT
         Uuid,
         ThreadId,
         AuthorId,
         Content
    FROM
        messages
    WHERE ThreadId = ?
    LIMIT ?, ?
    `, threadId.Bytes(), offset, count)

	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var m entities.Message
		err = rows.Scan(&m.Id, &m.ThreadId, &m.AuthorId, &m.Content)
		if err != nil {
			return err
		}
		appendToCollection(m)
	}
	err = rows.Err()
	return err
}

func GetMessageTotal(threadId uuid.UUID) (uint, error) {
	ret := uint(0)

	// TODO also need to put filtering in here
	err := db.QueryRow(`
    SELECT
        count(*) 
    FROM
        messages
    WHERE ThreadId = ?
    `, threadId.Bytes()).Scan(&ret)

	return ret, err
}
