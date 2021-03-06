package dbbackend

import (
	"database/sql"
	"github.com/john-sharp/jerver/entities"
	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"log"
	"strings"
)

var db *sql.DB
var getMessageStmt *sql.Stmt
var createMessageStmt *sql.Stmt
var deleteMessageStmt *sql.Stmt
var getThreadStmt *sql.Stmt
var createThreadStmt *sql.Stmt
var deleteThreadStmt *sql.Stmt
var editThreadStmt *sql.Stmt
var getUserByUnameStmt *sql.Stmt
var getUserByUuidStmt *sql.Stmt

func init() {
	var err error

	db, err = sql.Open("sqlite3", "./jerver.db")
	if err != nil {
		log.Fatal(err)
	}

	messagePrepareStmts()
	threadPrepareStmts()
	userPrepareStatements()
}

func messagePrepareStmts() {
	var err error
	getMessageStmt, err = db.Prepare(`
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

	deleteMessageStmt, err = db.Prepare(`
    DELETE FROM messages
    WHERE Uuid = ?
    `)

	if err != nil {
		log.Fatal(err)
	}
}

func threadPrepareStmts() {
	var err error

	getThreadStmt, err = db.Prepare(`
    SELECT 
         Uuid,
         Title
    FROM threads 
    WHERE Uuid = ?`)

	if err != nil {
		log.Fatal(err)
	}

	createThreadStmt, err = db.Prepare(`
    INSERT INTO threads (
        Uuid,
        Title )
    VALUES (?, ?)`)

	if err != nil {
		log.Fatal(err)
	}

	editThreadStmt, err = db.Prepare(`
    UPDATE threads SET Title=?
    WHERE Uuid = ?
    `)

	if err != nil {
		log.Fatal(err)
	}

	deleteThreadStmt, err = db.Prepare(`
    DELETE FROM threads
    WHERE Uuid = ?
    `)

	if err != nil {
		log.Fatal(err)
	}
}

func userPrepareStatements() {
	var err error
	getUserByUnameStmt, err = db.Prepare(`
    SELECT 
         Uuid,
         FirstName,
         SecondName,
         Username,
         HashedPwd
    FROM users 
    WHERE Username = ?`)

	if err != nil {
		log.Fatal(err)
	}

	getUserByUuidStmt, err = db.Prepare(`
    SELECT 
         Uuid,
         FirstName,
         SecondName,
         Username,
         HashedPwd
    FROM users 
    WHERE Uuid = ?`)

	if err != nil {
		log.Fatal(err)
	}
}

func GetMessageByUuid(targetUuid uuid.UUID) (*entities.Message, error) {
	var m entities.Message
	err := getMessageStmt.QueryRow(targetUuid.Bytes()).Scan(&m.Id, &m.ThreadId, &m.AuthorId, &m.Content)

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
	_, err := deleteMessageStmt.Exec(targetUuid.Bytes())
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

func EditMessageByUuid(targetUuid uuid.UUID, m *entities.MessageEdit) error {
	// TODO cache prepared update statements based on the
	// 'mask' of set fields
	// TODO can use reflect to loop through the fields of
	// the messageEdit struct to construct the query
	query := "UPDATE messages SET "
	updateFieldSql := []string{}
	params := []interface{}{}
	if m.ThreadId != nil {
		updateFieldSql = append(updateFieldSql, "ThreadId = ?")
		params = append(params, m.ThreadId.Bytes())
	}

	if m.AuthorId != nil {
		updateFieldSql = append(updateFieldSql, "AuthorId = ?")
		params = append(params, m.AuthorId.Bytes())
	}

	if m.Content != nil {
		updateFieldSql = append(updateFieldSql, "Content = ?")
		params = append(params, m.Content)
	}
	query += strings.Join(updateFieldSql, ", ")
	query += " WHERE Uuid = ?"
	params = append(params, targetUuid.Bytes())

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(params...)

	return err
}

func GetThreadByUuid(targetUuid uuid.UUID) (*entities.Thread, error) {
	var t entities.Thread
	err := getThreadStmt.QueryRow(targetUuid.Bytes()).Scan(&t.Id, &t.Title)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func CreateThread(t *entities.Thread) error {
	_, err := createThreadStmt.Exec(t.Id.Bytes(), t.Title)

	return err
}

func DeleteThreadByUuid(targetUuid uuid.UUID) error {
	_, err := deleteThreadStmt.Exec(targetUuid.Bytes())
	return err
}

func GetThreadCollection(count uint64, page int64, appendToCollection func(entities.Thread)) error {
	offset := page * int64(count)

	// TODO put in filtering
	rows, err := db.Query(`
    SELECT
        Uuid,
        Title
    FROM
        threads
    LIMIT ?, ?
    `, offset, count)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var t entities.Thread
		err = rows.Scan(&t.Id, &t.Title)
		if err != nil {
			log.Fatal(err)
		}
		appendToCollection(t)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	err = rows.Err()
	return err
}

func GetThreadTotal() (uint, error) {
	ret := uint(0)

	// TODO also need to put filtering in here
	err := db.QueryRow(`
    SELECT
        count(*) 
    FROM
        threads
    `).Scan(&ret)

	return ret, err
}

func EditThreadByUuid(targetUuid uuid.UUID, t *entities.ThreadEdit) error {
	_, err := editThreadStmt.Exec(t.Title, targetUuid.Bytes())
	return err
}

func GetUserByUsername(uname string) (*entities.User, error) {
	var u entities.User
	err := getUserByUnameStmt.QueryRow(uname).Scan(&u.Uuid, &u.FirstName, &u.SecondName, &u.Username, &u.HashedPwd)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func GetUserByUuid(targetUuid uuid.UUID) (*entities.User, error) {
	var u entities.User
	err := getUserByUuidStmt.QueryRow(targetUuid.Bytes()).Scan(&u.Uuid, &u.FirstName, &u.SecondName, &u.Username, &u.HashedPwd)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
