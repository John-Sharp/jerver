package main

import (
	"encoding/json"
	"errors"
	"github.com/john-sharp/jerver/entities"
	"github.com/satori/go.uuid"
	"techbrewers.com/usr/repos/entitycoll"
)

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

type thread entities.Thread
type threadNew thread

func (t *threadNew) UnmarshalJSON(b []byte) error {
	return (*thread)(t).verifyAndParseNew(b)
}

type threadCollection struct{}

var threads threadCollection

// implementation of entityCollectionInterface...

func (tc *threadCollection) GetRestName() string {
	return "threads"
}

func (tc *threadCollection) GetParentCollection() entitycoll.EntityCollection {
	return nil
}

func (tc *threadCollection) CreateEntity(requestor entitycoll.Entity, parentEntityUuids map[string]uuid.UUID, body []byte) (string, error) {
	var t threadNew
	err := json.Unmarshal(body, &t)
	if err != nil {
		return "", err
	}

	err = tc.create((*entities.Thread)(&t))
	path := "/" + tc.GetRestName() + "/" + t.Id.String()
	return path, nil
}

func (tc *threadCollection) GetEntity(targetUuid uuid.UUID) (entitycoll.Entity, error) {
	return tc.getByUuid(targetUuid)
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

	var err error
	ec.Entities, err = tc.getCollection(count, page)

	if err != nil {
		return entitycoll.Collection{}, err
	}

	ec.TotalEntities, err = tc.getTotal()

	if err != nil {
		return entitycoll.Collection{}, err
	}

	return ec, nil
}

func (tc *threadCollection) EditEntity(targetUuid uuid.UUID, body []byte) error {
	var edit entities.ThreadEdit

	err := json.Unmarshal(body, &edit)
	if err != nil {
		return err
	}

	if edit.Title == nil {
		return nil
	}

	return tc.editByUuid(targetUuid, &edit)
}

func (tc *threadCollection) DelEntity(targetUuid uuid.UUID) error {
	return tc.deleteByUuid(targetUuid)
}
