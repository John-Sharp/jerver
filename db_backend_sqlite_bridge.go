// +build SQLITE_BACKEND

package main

import (
	"github.com/john-sharp/jerver/entities"
	"github.com/john-sharp/jerver/sqlite-dbbackend"
	"github.com/satori/go.uuid"
	"techbrewers.com/usr/repos/entitycoll"
)

type messageCollection struct{}

func (mc *messageCollection) getByUuid(targetUuid uuid.UUID) (*entities.Message, error) {
	return dbbackend.GetMessageByUuid(targetUuid)
}

func (mc *messageCollection) create(m *entities.Message) error {
	return dbbackend.CreateMessage(m)
}

func (mc *messageCollection) deleteByUuid(targetUuid uuid.UUID) error {
	return dbbackend.DeleteMessageByUuid(targetUuid)
}

func (mc *messageCollection) getCollection(threadId uuid.UUID, count uint64, page int64) ([]entitycoll.Entity, error) {
	collection := []entitycoll.Entity{}

	messageCollectionAppender := func(m entities.Message) {
		collection = append(collection, m)
	}
	err := dbbackend.GetMessageCollection(threadId, count, page, messageCollectionAppender)

	if err != nil {
		collection = []entitycoll.Entity{}
	}
	return collection, err
}

func (mc *messageCollection) getTotal(threadId uuid.UUID) (uint, error) {
	return dbbackend.GetMessageTotal(threadId)
}
