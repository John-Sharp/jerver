package main

import (
	"encoding/json"
	"errors"
	"github.com/john-sharp/jerver/entities"
	"github.com/satori/go.uuid"
	"gitlab.com/johncolinsharp/entitycoll"
)

type message entities.Message

var messages messageCollection

// implementation of entityCollectionInterface...

func (mc *messageCollection) GetRestName() string {
	return "messages"
}

func (mc *messageCollection) GetParentCollection() entitycoll.APINode {
	return &threads
}

func (mc *messageCollection) CreateEntity(requestor entitycoll.Entity, parentEntityUuids map[string]uuid.UUID, body []byte) (string, error) {
	var m entities.Message

	threadId, ok := parentEntityUuids["threads"]
	if !ok {
		return "", errors.New("no thread ID supplied")
	}

	err := json.Unmarshal(body, &m)
	if err != nil {
		return "", err
	}

	m.Id, _ = uuid.NewV4()
	m.ThreadId = threadId
	user := requestor.(*user)
	m.AuthorId = user.Uuid

	err = mc.create(&m)

	if err != nil {
		return "", err
	}

	path := "/" + mc.GetParentCollection().GetRestName() + "/" + threadId.String() + "/" + mc.GetRestName() + "/" + m.Id.String()
	return path, nil
}

func (mc *messageCollection) GetEntity(requestor entitycoll.Entity, targetUuid uuid.UUID) (entitycoll.Entity, error) {
	return mc.getByUuid(targetUuid)
}

func (mc *messageCollection) GetCollection(requestor entitycoll.Entity, parentEntityUuids map[string]uuid.UUID, filter entitycoll.CollFilter) (entitycoll.Collection, error) {
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

	var err error
	ec.Entities, err = mc.getCollection(threadId, count, page)

	if err != nil {
		return entitycoll.Collection{}, err
	}

	ec.TotalEntities, err = mc.getTotal(threadId)

	if err != nil {
		return entitycoll.Collection{}, err
	}

	return ec, nil
}

func (mc *messageCollection) EditEntity(requestor entitycoll.Entity, targetUuid uuid.UUID, body []byte) error {
	var edit entities.MessageEdit

	err := json.Unmarshal(body, &edit)
	if err != nil {
		return err
	}

	if edit.ThreadId == nil && edit.AuthorId == nil && edit.Content == nil {
		return nil
	}

	return mc.editByUuid(targetUuid, &edit)
}

func (mc *messageCollection) DelEntity(requestor entitycoll.Entity, targetUuid uuid.UUID) error {
	return mc.deleteByUuid(targetUuid)
}
