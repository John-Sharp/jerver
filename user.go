package main

import (
	"errors"
	"github.com/john-sharp/jerver/entities"
	"github.com/satori/go.uuid"
	"gitlab.com/johncolinsharp/entitycoll"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type user entities.User

func (u *user) popNew(fname, sname, uname, pwd string) error {
	hpwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	u.Uuid, _ = uuid.NewV4()
	u.FirstName = fname
	u.SecondName = sname
	u.Username = uname
	u.HashedPwd = hpwd

	return nil
}

func (uc *userCollection) verifyUser(uname, pwd string) (entitycoll.Entity, error) {
	u, err := uc.getUserByUsername(uname)

	if err != nil {
		log.Printf(err.Error())
		return nil, errors.New("could not find user")
	}

	if err = bcrypt.CompareHashAndPassword(u.HashedPwd, []byte(pwd)); err != nil {
		return nil, err
	}

	return u, nil
}

// userCollection will implement entityCollection
type userCollection struct{}

// global variable representing userCollection of all users in example
var users userCollection

// implementation of entityCollectionInterface...

func (uc *userCollection) GetRestName() string {
	return "users"
}

func (uc *userCollection) GetParentCollection() entitycoll.APINode {
	return nil
}

func (uc *userCollection) CreateEntity(requestor entitycoll.Entity, parentEntityUuids map[string]uuid.UUID, body []byte) (string, error) {
	// TODO add dedicated error code to entity coll for
	// not being able to create entity
	return "", errors.New("create entity not allowed")
}

func (uc *userCollection) GetEntity(requestor entitycoll.Entity, targetUuid uuid.UUID) (entitycoll.Entity, error) {
	return uc.getUserByUuid(targetUuid)
}

func (uc *userCollection) GetCollection(requestor entitycoll.Entity, parentEntityUuids map[string]uuid.UUID, filter entitycoll.CollFilter) (entitycoll.Collection, error) {
	return entitycoll.Collection{}, nil
}

func (uc *userCollection) EditEntity(requestor entitycoll.Entity, targetUuid uuid.UUID, body []byte) error {
	return nil
}

func (uc *userCollection) DelEntity(requestor entitycoll.Entity, targetUuid uuid.UUID) error {
	return errors.New("del entity not allowed")
}
