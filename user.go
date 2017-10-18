package main

import (
	"encoding/json"
	"errors"
	"github.com/john-sharp/entitycoll"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// user is an api entity
type user struct {
	Uuid       uuid.UUID
	FirstName  string
	SecondName string
	Username   string
	HashedPwd  []byte
}

// called during the jsonUnmarshal of a new user.
// Do verification of contents of request body
// and additional creation processes here (such
// as generating a unique id)
func (u *user) verifyAndParseNew(b []byte) error {
	var t struct {
		FirstName  *string
		SecondName *string
	}
	err := json.Unmarshal(b, &t)

	if err != nil {
		return err
	}

	if t.FirstName == nil {
		return errors.New("user FirstName not set when required")
	}

	if t.SecondName == nil {
		return errors.New("user SecondName not set when required")
	}

	u.Uuid = uuid.NewV4()
	u.FirstName = *t.FirstName
	u.SecondName = *t.SecondName
	return nil
}

func (u *user) popNew(fname, sname, uname, pwd string) error {
	hpwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	u.Uuid = uuid.NewV4()
	u.FirstName = fname
	u.SecondName = sname
	u.Username = uname
	u.HashedPwd = hpwd

	return nil
}

func (uc *userCollection) verifyUser(uname, pwd string) (entitycoll.Entity, error) {
	var i int
	for i, _ = range *uc {
		if (*uc)[i].Username == uname {
			if err := bcrypt.CompareHashAndPassword((*uc)[i].HashedPwd, []byte(pwd)); err != nil {
				return nil, err
			} else {
				return (&(*uc)[i]), nil
			}
		}
	}
	return nil, errors.New("could not find user")
}

// called during the jsonUnmarshal of an edit of a user.
// Do error checking here, and make sure that any fields
// that need to be preserved are preserved (in this case
// the Uuid)
func (u *user) verifyAndParseEdit(b []byte) error {
	bu := u.Uuid
	err := json.Unmarshal(b, &u)
	if err != nil {
		return err
	}
	u.Uuid = bu
	return nil
}

// userNew and userEdit are types so that parsing of JSON
// is done differently for new users to how it is done for
// edited users
type userNew user

// defining this method means that 'verifyAndParseNew' is
// called whenever the JSON parser encounters a userNew type
func (u *userNew) UnmarshalJSON(b []byte) error {
	return (*user)(u).verifyAndParseNew(b)
}

type userEdit user

func (u *userEdit) UnmarshalJSON(b []byte) error {
	return (*user)(u).verifyAndParseEdit(b)
}

// userCollection will implement entityCollection
// in this example it is just defined as a slice of users
// in future could be a wrapper around db connection etc
type userCollection []user

// global variable representing userCollection of all users in example
var users userCollection

// implementation of entityCollectionInterface...

func (uc *userCollection) getRestName() string {
	return "users"
}

func (uc *userCollection) getParentCollection() entityCollection {
	return nil
}

func (uc *userCollection) createEntity(creatingUser *user, parentEntityUuids map[string]uuid.UUID, body []byte) (string, error) {
	var u user
	err := json.Unmarshal(body, (*userNew)(&u))
	if err != nil {
		return "", err
	}
	*uc = append(*uc, u)
	path := "/" + uc.getRestName() + "/" + u.Uuid.String()
	return path, nil
}

func (uc *userCollection) getEntity(targetUuid uuid.UUID) (entitycoll.Entity, error) {
	var i int
	for i, _ = range *uc {
		if uuid.Equal((*uc)[i].Uuid, targetUuid) {
			return &(*uc)[i], nil
		}
	}
	return nil, errors.New("could not find user")
}

func (uc *userCollection) getCollection(parentEntityUuids map[string]uuid.UUID, filter collFilter) (collection, error) {
	return collection{TotalEntities: 0, Entities: uc}, nil
}

func (uc *userCollection) editEntity(targetUuid uuid.UUID, body []byte) error {
	e, err := uc.getEntity(targetUuid)
	if err != nil {
		return err
	}

	u, ok := e.(*user)
	if !ok {
		return errors.New("user pointer type assertion error")
	}

	return json.Unmarshal(body, (*userEdit)(u))
}

func (uc *userCollection) delEntity(targetUuid uuid.UUID) error {
	var i int
	for i, _ = range *uc {
		if uuid.Equal((*uc)[i].Uuid, targetUuid) {
			(*uc) = append((*uc)[:i], (*uc)[i+1:]...)
			return nil
		}
	}
	return errors.New("could not find user to delete")
}
