package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"techbrewers.com/usr/repos/entitycoll"
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

	u.Uuid, _ = uuid.NewV4()
	u.FirstName = *t.FirstName
	u.SecondName = *t.SecondName
	return nil
}

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

func (uc *userCollection) getUserByUsername(uname string) (*user, error) {
	var u user
	err := uc.getFromUnameStmt.QueryRow(uname).Scan(&u.Uuid, &u.FirstName, &u.SecondName, &u.Username, &u.HashedPwd)

	if err != nil {
		return nil, err
	}

	return &u, nil
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
type userCollection struct {
	getFromUnameStmt *sql.Stmt
	getFromUuidStmt  *sql.Stmt
}

// global variable representing userCollection of all users in example
var users userCollection

func (uc *userCollection) prepareStmts() {
	var err error
	uc.getFromUnameStmt, err = db.Prepare(`
    SELECT 
         Uuid,
         FirstName,
         SecondName,
         Username,
         HashedPwd
    FROM users 
    WHERE Username = ?`)

	if err != nil {
		log.Fatal(err)
	}

	uc.getFromUuidStmt, err = db.Prepare(`
    SELECT 
         Uuid,
         FirstName,
         SecondName,
         Username,
         HashedPwd
    FROM users 
    WHERE Uuid = ?`)

	if err != nil {
		log.Fatal(err)
	}
}

func (uc *userCollection) closeStmts() {
	users.getFromUnameStmt.Close()
	users.getFromUuidStmt.Close()
}

// implementation of entityCollectionInterface...

func (uc *userCollection) GetRestName() string {
	return "users"
}

func (uc *userCollection) GetParentCollection() entitycoll.EntityCollection {
	return nil
}

func (uc *userCollection) CreateEntity(requestor entitycoll.Entity, parentEntityUuids map[string]uuid.UUID, body []byte) (string, error) {
	// TODO add dedicated error code to entity coll for
	// not being able to create entity
	return "", errors.New("create entity not allowed")
}

func (uc *userCollection) GetEntity(targetUuid uuid.UUID) (entitycoll.Entity, error) {
	var u user
	err := uc.getFromUuidStmt.QueryRow(targetUuid.Bytes()).Scan(&u.Uuid, &u.FirstName, &u.SecondName, &u.Username, &u.HashedPwd)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &u, nil
}

func (uc *userCollection) GetCollection(parentEntityUuids map[string]uuid.UUID, filter entitycoll.CollFilter) (entitycoll.Collection, error) {
	return entitycoll.Collection{}, nil
}

func (uc *userCollection) EditEntity(targetUuid uuid.UUID, body []byte) error {
	e, err := uc.GetEntity(targetUuid)
	if err != nil {
		return err
	}

	u, ok := e.(*user)
	if !ok {
		return errors.New("user pointer type assertion error")
	}

	return json.Unmarshal(body, (*userEdit)(u))
}

func (uc *userCollection) DelEntity(targetUuid uuid.UUID) error {
	return errors.New("del entity not allowed")
}
