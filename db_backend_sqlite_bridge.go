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

func (mc *messageCollection) editByUuid(targetUuid uuid.UUID, m *entities.MessageEdit) error {
	return dbbackend.EditMessageByUuid(targetUuid, m)
}

func (tc *threadCollection) getByUuid(targetUuid uuid.UUID) (*entities.Thread, error) {
	return dbbackend.GetThreadByUuid(targetUuid)
}

func (tc *threadCollection) create(t *entities.Thread) error {
	return dbbackend.CreateThread(t)
}

func (tc *threadCollection) deleteByUuid(targetUuid uuid.UUID) error {
	return dbbackend.DeleteThreadByUuid(targetUuid)
}

func (mc *threadCollection) getCollection(count uint64, page int64) ([]entitycoll.Entity, error) {
	collection := []entitycoll.Entity{}

	threadCollectionAppender := func(t entities.Thread) {
		collection = append(collection, t)
	}
	err := dbbackend.GetThreadCollection(count, page, threadCollectionAppender)

	if err != nil {
		collection = []entitycoll.Entity{}
	}
	return collection, err
}

func (tc *threadCollection) getTotal() (uint, error) {
	return dbbackend.GetThreadTotal()
}

func (tc *threadCollection) editByUuid(targetUuid uuid.UUID, t *entities.ThreadEdit) error {
	return dbbackend.EditThreadByUuid(targetUuid, t)
}
