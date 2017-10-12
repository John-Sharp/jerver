package main

import (
	"encoding/json"
	"errors"
	"github.com/satori/go.uuid"
)

type thread struct {
	Id      uuid.UUID
	Title   string
	NumMsgs uint
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

func (t *thread) verifyAndParseEdit(b []byte) error {
	bu := t.Id
	bn := t.NumMsgs
	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	t.Id = bu
	t.NumMsgs = bn
	return nil
}

type threadNew thread

func (t *threadNew) UnmarshalJSON(b []byte) error {
	return (*thread)(t).verifyAndParseNew(b)
}

type threadEdit thread

func (t *threadEdit) UnmarshalJSON(b []byte) error {
	return (*thread)(t).verifyAndParseEdit(b)
}

type threadCollection []thread

var threads threadCollection

// implementation of entityCollectionInterface...

func (tc *threadCollection) getRestName() string {
	return "threads"
}

func (tc *threadCollection) getParentCollection() entityCollection {
	return nil
}

func (tc *threadCollection) createEntity(user *user, parentEntityUuids map[string]uuid.UUID, body []byte) (string, error) {
	var t thread
	err := json.Unmarshal(body, (*threadNew)(&t))
	if err != nil {
		return "", err
	}
	*tc = append(*tc, t)
	path := "/" + tc.getRestName() + "/" + t.Id.String()
	return path, nil
}

func (tc *threadCollection) getEntity(targetUuid uuid.UUID) (entity, error) {
	var i int
	for i, _ = range *tc {
		if uuid.Equal((*tc)[i].Id, targetUuid) {
			return &(*tc)[i], nil
		}
	}
	return nil, errors.New("could not find user")
}

func (tc *threadCollection) getCollection(parentEntityUuids map[string]uuid.UUID) (interface{}, error) {
	return tc, nil
}

func (tc *threadCollection) editEntity(targetUuid uuid.UUID, body []byte) error {
	e, err := tc.getEntity(targetUuid)
	if err != nil {
		return err
	}

	u, ok := e.(*thread)
	if !ok {
		return errors.New("thread pointer type assertion error")
	}

	return json.Unmarshal(body, (*threadEdit)(u))
}

func (tc *threadCollection) delEntity(targetUuid uuid.UUID) error {
	var i int
	for i, _ = range *tc {
		if uuid.Equal((*tc)[i].Id, targetUuid) {
			(*tc) = append((*tc)[:i], (*tc)[i+1:]...)
			return nil
		}
	}
	return errors.New("could not find thread to delete")
}
