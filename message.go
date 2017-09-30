package main

import (
	"encoding/json"
	"errors"
	"github.com/satori/go.uuid"
)

type message struct {
	Id       uuid.UUID
	ThreadId uuid.UUID
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

type messageCollection []message

var messages messageCollection

// implementation of entityCollectionInterface...

func (mc *messageCollection) createEntity(parentEntityUuids map[string]uuid.UUID, body []byte) error {
	var m message

	threadId, ok := parentEntityUuids["threads"]
	if !ok {
		return errors.New("no thread ID supplied")
	}

	err := json.Unmarshal(body, (*messageNew)(&m))
	if err != nil {
		return err
	}

	m.ThreadId = threadId

	*mc = append(*mc, m)
	return nil
}

func (mc *messageCollection) getEntity(targetUuid uuid.UUID) (entity, error) {
	var i int
	for i, _ = range *mc {
		if uuid.Equal((*mc)[i].Id, targetUuid) {
			return &(*mc)[i], nil
		}
	}
	return nil, errors.New("could not find message")
}

func (mc *messageCollection) getCollection(parentEntityUuids map[string]uuid.UUID) (interface{}, error) {
	threadId, ok := parentEntityUuids["threads"]
	if !ok {
		return nil, errors.New("no thread ID supplied")
	}

	var mSubColl []message = []message{}

	for _, m := range *mc {
		if uuid.Equal(m.ThreadId, threadId) {
			mSubColl = append(mSubColl, m)
		}
	}

	return mSubColl, nil
}

func (mc *messageCollection) editEntity(targetUuid uuid.UUID, body []byte) error {
	e, err := mc.getEntity(targetUuid)
	if err != nil {
		return err
	}

	m, ok := e.(*message)
	if !ok {
		return errors.New("message pointer type assertion error")
	}

	return json.Unmarshal(body, (*messageEdit)(m))
}

func (mc *messageCollection) delEntity(targetUuid uuid.UUID) error {
	var i int
	for i, _ = range *mc {
		if uuid.Equal((*mc)[i].Id, targetUuid) {
			(*mc) = append((*mc)[:i], (*mc)[i+1:]...)
			return nil
		}
	}
	return errors.New("could not find message to delete")
}
