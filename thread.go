package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"log"
	"techbrewers.com/usr/repos/entitycoll"
)

type thread struct {
	Id      uuid.UUID
	Title   string
	NumMsgs uint
}

type threadEdit struct {
	Title *string
}

func (t *thread) verifyAndParseNew(b []byte) error {
	var data struct {
		Title *string
	}

	err := json.Unmarshal(b, &data)

	if err != nil {
		return err
	}

	if data.Title == nil {
		return errors.New("thread Title not set when required")
	}

	t.Id, _ = uuid.NewV4()
	t.Title = *data.Title
	return nil
}

type threadNew thread

func (t *threadNew) UnmarshalJSON(b []byte) error {
	return (*thread)(t).verifyAndParseNew(b)
}

type threadCollection struct {
	getFromUuidStmt  *sql.Stmt
	createEntityStmt *sql.Stmt
	editEntityStmt   *sql.Stmt
	deleteEntityStmt *sql.Stmt
}

func (tc *threadCollection) prepareStmts() {
	var err error

	tc.getFromUuidStmt, err = db.Prepare(`
    SELECT 
         Uuid,
         Title
    FROM threads 
    WHERE Uuid = ?`)

	if err != nil {
		log.Fatal(err)
	}

	tc.createEntityStmt, err = db.Prepare(`
    INSERT INTO threads (
        Uuid,
        Title )
    VALUES (?, ?)`)

	if err != nil {
		log.Fatal(err)
	}

	tc.editEntityStmt, err = db.Prepare(`
    UPDATE threads SET Title=?
    WHERE Uuid = ?
    `)

	if err != nil {
		log.Fatal(err)
	}

	tc.deleteEntityStmt, err = db.Prepare(`
    DELETE FROM threads
    WHERE Uuid = ?
    `)

	if err != nil {
		log.Fatal(err)
	}
}

func (tc *threadCollection) closeStmts() {
	tc.getFromUuidStmt.Close()
}

var threads threadCollection

// implementation of entityCollectionInterface...

func (tc *threadCollection) GetRestName() string {
	return "threads"
}

func (tc *threadCollection) GetParentCollection() entitycoll.EntityCollection {
	return nil
}

func (tc *threadCollection) CreateEntity(requestor entitycoll.Entity, parentEntityUuids map[string]uuid.UUID, body []byte) (string, error) {
	var t thread
	err := json.Unmarshal(body, (*threadNew)(&t))
	if err != nil {
		return "", err
	}

	_, err = tc.createEntityStmt.Exec(t.Id.Bytes(), t.Title)

	if err != nil {
		return "", err
	}

	path := "/" + tc.GetRestName() + "/" + t.Id.String()
	return path, nil
}

func (tc *threadCollection) GetEntity(targetUuid uuid.UUID) (entitycoll.Entity, error) {
	var t thread
	err := tc.getFromUuidStmt.QueryRow(targetUuid.Bytes()).Scan(&t.Id, &t.Title)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (tc *threadCollection) GetCollection(parentEntityUuids map[string]uuid.UUID, filter entitycoll.CollFilter) (entitycoll.Collection, error) {
	var ec entitycoll.Collection

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
        Title
    FROM
        threads
    LIMIT ?, ?
    `, offset, count)
	if err != nil {
		return entitycoll.Collection{}, err
	}
	defer rows.Close()
	for rows.Next() {
		ec.TotalEntities += 1
		var t thread
		err = rows.Scan(&t.Id, &t.Title)
		if err != nil {
			log.Fatal(err)
		}
		ec.Entities = append(ec.Entities, t)
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
        threads
    `).Scan(&ec.TotalEntities)
	if err != nil {
		return entitycoll.Collection{}, err
	}

	return ec, nil
}

func (tc *threadCollection) EditEntity(targetUuid uuid.UUID, body []byte) error {
	var edit threadEdit

	err := json.Unmarshal(body, &edit)
	if err != nil {
		return err
	}

	if edit.Title == nil {
		return nil
	}

	_, err = tc.editEntityStmt.Exec(edit.Title, targetUuid.Bytes())

	return err
}

func (tc *threadCollection) DelEntity(targetUuid uuid.UUID) error {
	_, err := tc.deleteEntityStmt.Exec(targetUuid.Bytes())
	return err
}
