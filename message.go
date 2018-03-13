package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"log"
	"strings"
	"techbrewers.com/usr/repos/entitycoll"
)

type message struct {
	Id       uuid.UUID
	ThreadId uuid.UUID
	AuthorId uuid.UUID
	Content  string
}

type messageEdit struct {
	ThreadId *uuid.UUID
	AuthorId *uuid.UUID
	Content  *string
}

func (m *message) verifyAndParseNew(b []byte) error {
	err := json.Unmarshal(b, m)

	if err != nil {
		return err
	}

	m.Id, _ = uuid.NewV4()
	return nil
}

func (m *message) verifyAndParseEdit(b []byte) error {
	bu := m.Id
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	m.Id = bu
	return nil
}

type messageNew message

func (m *messageNew) UnmarshalJSON(b []byte) error {
	return (*message)(m).verifyAndParseNew(b)
}

type messageCollection struct {
	getFromUuidStmt  *sql.Stmt
	createEntityStmt *sql.Stmt
	deleteEntityStmt *sql.Stmt
}

func (mc *messageCollection) prepareStmts() {
	var err error

	mc.getFromUuidStmt, err = db.Prepare(`
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

	mc.createEntityStmt, err = db.Prepare(`
    INSERT INTO messages (
        Uuid,
        ThreadId,
        AuthorId,
        Content)
    VALUES (?, ?, ?, ?)`)

	if err != nil {
		log.Fatal(err)
	}

	mc.deleteEntityStmt, err = db.Prepare(`
    DELETE FROM messages
    WHERE Uuid = ?
    `)

	if err != nil {
		log.Fatal(err)
	}
}

var messages messageCollection

// implementation of entityCollectionInterface...

func (mc *messageCollection) GetRestName() string {
	return "messages"
}

func (mc *messageCollection) GetParentCollection() entitycoll.EntityCollection {
	return &threads
}

func (mc *messageCollection) CreateEntity(requestor entitycoll.Entity, parentEntityUuids map[string]uuid.UUID, body []byte) (string, error) {
	var m message

	threadId, ok := parentEntityUuids["threads"]
	if !ok {
		return "", errors.New("no thread ID supplied")
	}

	err := json.Unmarshal(body, (*messageNew)(&m))
	if err != nil {
		return "", err
	}

	m.ThreadId = threadId
	user := requestor.(*user)
	m.AuthorId = user.Uuid

	_, err = mc.createEntityStmt.Exec(
		m.Id.Bytes(),
		m.ThreadId.Bytes(),
		m.AuthorId.Bytes(),
		m.Content)

	path := "/" + mc.GetParentCollection().GetRestName() + "/" + threadId.String() + "/" + mc.GetRestName() + "/" + m.Id.String()
	return path, nil
}

func (mc *messageCollection) GetEntity(targetUuid uuid.UUID) (entitycoll.Entity, error) {
	var m message
	err := mc.getFromUuidStmt.QueryRow(targetUuid.Bytes()).Scan(&m.Id, &m.ThreadId, &m.AuthorId, &m.Content)

	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (mc *messageCollection) GetCollection(parentEntityUuids map[string]uuid.UUID, filter entitycoll.CollFilter) (entitycoll.Collection, error) {
	var ec entitycoll.Collection
	threadId, ok := parentEntityUuids["threads"]
	if !ok {
		return entitycoll.Collection{}, errors.New("no thread ID supplied")
	}

	count := uint64(10)
	page := int64(0)
	if filter.Page != nil {
		page = *filter.Page
	}
	if filter.Count != nil {
		count = *filter.Count
	}
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
		return entitycoll.Collection{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var m message
		err = rows.Scan(&m.Id, &m.ThreadId, &m.AuthorId, &m.Content)
		if err != nil {
			return entitycoll.Collection{}, err
		}
		ec.Entities = append(ec.Entities, m)
	}
	err = rows.Err()
	if err != nil {
		return entitycoll.Collection{}, err
	}

	// TODO also need to put filtering in here
	err = db.QueryRow(`
    SELECT
        count(*) 
    FROM
        messages
    WHERE ThreadId = ?
    `, threadId.Bytes()).Scan(&ec.TotalEntities)
	if err != nil {
		return entitycoll.Collection{}, err
	}

	return ec, nil
}

func (mc *messageCollection) EditEntity(targetUuid uuid.UUID, body []byte) error {
	var edit messageEdit

	err := json.Unmarshal(body, &edit)
	if err != nil {
		return err
	}

	if edit.ThreadId == nil && edit.AuthorId == nil && edit.Content == nil {
		return nil
	}

	// TODO cache prepared update statements based on the
	// 'mask' of set fields
	// TODO can use reflect to loop through the fields of
	// the messageEdit struct to construct the query
	query := "UPDATE messages SET "
	updateFieldSql := []string{}
	params := []interface{}{}
	if edit.ThreadId != nil {
		updateFieldSql = append(updateFieldSql, "ThreadId = ?")
		params = append(params, edit.ThreadId.Bytes())
	}

	if edit.AuthorId != nil {
		updateFieldSql = append(updateFieldSql, "AuthorId = ?")
		params = append(params, edit.AuthorId.Bytes())
	}

	if edit.Content != nil {
		updateFieldSql = append(updateFieldSql, "Content = ?")
		params = append(params, edit.Content)
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

func (mc *messageCollection) DelEntity(targetUuid uuid.UUID) error {
	_, err := mc.deleteEntityStmt.Exec(targetUuid.Bytes())
	return err
}
