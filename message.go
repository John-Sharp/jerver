package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/john-sharp/entitycoll"
	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"log"
)

type message struct {
	Id       uuid.UUID
	ThreadId uuid.UUID
	AuthorId uuid.UUID
	Content  string
}

func (m *message) verifyAndParseNew(b []byte) error {
	err := json.Unmarshal(b, m)

	if err != nil {
		return err
	}

	m.Id = uuid.NewV4()
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

type messageEdit message

func (m *messageEdit) UnmarshalJSON(b []byte) error {
	return (*message)(m).verifyAndParseEdit(b)
}

type messageCollection struct {
	messages        []message
	getFromUuidStmt *sql.Stmt
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

	// increase parent thread's number of messages by one
	threadEntity, err := threads.GetEntity(threadId)
	if err != nil {
		return "", err
	}
	threadEntity.(*thread).NumMsgs += 1

	mc.messages = append(mc.messages, m)
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

	// TODO put in paging and filtering
	rows, err := db.Query(`
    SELECT
         Uuid,
         ThreadId,
         AuthorId,
         Content
    FROM
        messages
    WHERE ThreadId = ?
    `, threadId.Bytes())

	if err != nil {
		return entitycoll.Collection{}, err
	}
	defer rows.Close()
	for rows.Next() {
		ec.TotalEntities += 1
		var m message
		err = rows.Scan(&m.Id, &m.ThreadId, &m.AuthorId, &m.Content)
		if err != nil {
			log.Fatal(err)
		}
		ec.Entities = append(ec.Entities, m)
	}
	err = rows.Err()
	if err != nil {
		return entitycoll.Collection{}, err
	}

	return ec, nil

	// count := uint64(10)
	// page := int64(0)
	// if filter.Page != nil {
	// 	page = *filter.Page
	// }
	// if filter.Count != nil {
	// 	count = *filter.Count
	// }
	// offset := page * int64(count)

	// i := uint(0)
	// for _, m := range mc.messages {
	// 	if uuid.Equal(m.ThreadId, threadId) {
	// 		if int64(i) >= int64(count)+offset {
	// 			break
	// 		}
	// 		if int64(i) >= offset {
	// 			mSubColl = append(mSubColl, m)
	// 		}
	// 		i++
	// 	}
	// }

	// return entitycoll.Collection{TotalEntities: i /*, Entities: mSubColl*/}, nil
}

func (mc *messageCollection) EditEntity(targetUuid uuid.UUID, body []byte) error {
	e, err := mc.GetEntity(targetUuid)
	if err != nil {
		return err
	}

	m, ok := e.(*message)
	if !ok {
		return errors.New("message pointer type assertion error")
	}

	return json.Unmarshal(body, (*messageEdit)(m))
}

func (mc *messageCollection) DelEntity(targetUuid uuid.UUID) error {
	var i int
	for i, _ = range mc.messages {
		if uuid.Equal(mc.messages[i].Id, targetUuid) {
			threadEntity, err := threads.GetEntity(mc.messages[i].ThreadId)
			if err != nil {
				fmt.Println("WARNING: Could not find parent thread when deleting")
			} else {
				threadEntity.(*thread).NumMsgs -= 1
			}

			mc.messages = append(mc.messages[:i], mc.messages[i+1:]...)
			return nil
		}
	}
	return errors.New("could not find message to delete")
}
