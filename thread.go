package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"encoding/json"
	"errors"
	"github.com/john-sharp/entitycoll"
	"github.com/satori/go.uuid"
    "log"
)

type thread struct {
	Id      uuid.UUID
	Title   string
	NumMsgs uint
}

type threadEdit struct {
    Title  *string
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

	t.Id = uuid.NewV4()
	t.Title = *data.Title
	return nil
}

// func (t *thread) verifyAndParseEdit(b []byte) error {
//	bu := t.Id
//	bn := t.NumMsgs
//	err := json.Unmarshal(b, &t)
//	if err != nil {
//		return err
//	}
//	t.Id = bu
//	t.NumMsgs = bn
//	return nil
// }

type threadNew thread

func (t *threadNew) UnmarshalJSON(b []byte) error {
	return (*thread)(t).verifyAndParseNew(b)
}

// type threadEdit thread

//func (t *threadEdit) UnmarshalJSON(b []byte) error {
//	return (*thread)(t).verifyAndParseEdit(b)
//}

type threadCollection struct {
    threads []thread
    getFromUuidStmt *sql.Stmt
    createEntityStmt *sql.Stmt
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

    // TODO put in paging and filtering
    rows, err := db.Query(`
    SELECT
        Uuid,
        Title
    FROM
        threads
    `)
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
	// count := uint64(10)
	// page := int64(0)
	// if filter.Page != nil {
	// 	page = *filter.Page
	// }
	// if filter.Count != nil {
	// 	count = *filter.Count
	// }
	// offset := page * int64(count)

    return ec, nil
}

func (tc *threadCollection) EditEntity(targetUuid uuid.UUID, body []byte) error {
    var edit threadEdit

    err := json.Unmarshal(body, &edit)
    if err != nil {
        return err
    }

    if (edit.Title != nil) {
        log.Println(*edit.Title)
    }

    return err
	// e, err := tc.GetEntity(targetUuid)
	// if err != nil {
	// 	return err
	// }

	// u, ok := e.(*thread)
	// if !ok {
	// 	return errors.New("thread pointer type assertion error")
	// }

	// return json.Unmarshal(body, (*threadEdit)(u))
}

func (tc *threadCollection) DelEntity(targetUuid uuid.UUID) error {
	var i int
	for i, _ = range tc.threads {
		if uuid.Equal(tc.threads[i].Id, targetUuid) {
			tc.threads = append(tc.threads[:i], tc.threads[i+1:]...)
			return nil
		}
	}
	return errors.New("could not find thread to delete")
}
