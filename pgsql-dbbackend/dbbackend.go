package dbbackend

import (
	"database/sql"
	"fmt"
	"github.com/john-sharp/jerver/entities"
	_ "github.com/lib/pq"
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

	connStr := "user=jerver dbname=jerver sslmode=disable"
	db, err = sql.Open("postgres", connStr)
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
	WHERE Uuid = $1`)

	if err != nil {
		log.Fatal(err)
	}

	createMessageStmt, err = db.Prepare(`
    INSERT INTO messages (
        Uuid,
        ThreadId,
        AuthorId,
        Content)
    VALUES ($1, $2, $3, $4)`)

	if err != nil {
		log.Fatal(err)
	}

	deleteMessageStmt, err = db.Prepare(`
    DELETE FROM messages
    WHERE Uuid = $1
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
    WHERE Uuid = $1`)

	if err != nil {
		log.Fatal(err)
	}

	createThreadStmt, err = db.Prepare(`
    INSERT INTO threads (
        Uuid,
        Title )
    VALUES ($1, $2)`)

	if err != nil {
		log.Fatal(err)
	}

	editThreadStmt, err = db.Prepare(`
    UPDATE threads SET Title=$1
    WHERE Uuid = $2
    `)

	if err != nil {
		log.Fatal(err)
	}

	deleteThreadStmt, err = db.Prepare(`
    DELETE FROM threads
    WHERE Uuid = $3
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
    WHERE Username = $1`)

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
    WHERE Uuid = $1`)

	if err != nil {
		log.Fatal(err)
	}
}

func GetMessageByUuid(targetUuid uuid.UUID) (*entities.Message, error) {
	var m entities.Message
	err := getMessageStmt.QueryRow(targetUuid).Scan(&m.Id, &m.ThreadId, &m.AuthorId, &m.Content)

	if err != nil {
		return nil, err
	}
	return &m, nil
}

func CreateMessage(m *entities.Message) error {
	_, err := createMessageStmt.Exec(
		m.Id,
		m.ThreadId,
		m.AuthorId,
		m.Content)

	return err
}

func DeleteMessageByUuid(targetUuid uuid.UUID) error {
	_, err := deleteMessageStmt.Exec(targetUuid)
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
    WHERE ThreadId = $1
    LIMIT $2, $3
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
    WHERE ThreadId = $1
    `, threadId).Scan(&ret)

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
	var paramIndex = 1
	if m.ThreadId != nil {
		updateFieldSql = append(updateFieldSql, fmt.Sprintf("ThreadId = $%d", paramIndex))
		paramIndex += 1
		params = append(params, m.ThreadId.Bytes())
	}

	if m.AuthorId != nil {
		updateFieldSql = append(updateFieldSql, fmt.Sprintf("AuthorId = $%d", paramIndex))
		paramIndex += 1
		params = append(params, m.AuthorId.Bytes())
	}

	if m.Content != nil {
		updateFieldSql = append(updateFieldSql, fmt.Sprintf("Content = $%d", paramIndex))
		paramIndex += 1
		params = append(params, m.Content)
	}
	query += strings.Join(updateFieldSql, ", ")
	query += fmt.Sprintf(" WHERE Uuid = $%d", paramIndex)
	paramIndex += 1
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
	err := getThreadStmt.QueryRow(targetUuid).Scan(&t.Id, &t.Title)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func CreateThread(t *entities.Thread) error {
	_, err := createThreadStmt.Exec(t.Id, t.Title)

	return err
}

func DeleteThreadByUuid(targetUuid uuid.UUID) error {
	_, err := deleteThreadStmt.Exec(targetUuid)
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
    LIMIT $1, $2
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
	_, err := editThreadStmt.Exec(t.Title, targetUuid)
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
	err := getUserByUuidStmt.QueryRow(targetUuid).Scan(&u.Uuid, &u.FirstName, &u.SecondName, &u.Username, &u.HashedPwd)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
