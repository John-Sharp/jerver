package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
)

// interface for generic collection of api entities
type entityCollection interface {

	// given a []byte containing JSON, should create an entity and
	// add it to the collection
	createEntity(body []byte) error

	// given a Uuid should find entity in collection and return
	getEntity(targetUuid uuid.UUID) (entity, error)

	// return whole collection (in future maybe include filters
	// in argument and return subset of collection)
	getCollection() (interface{}, error)

	// edit entity with Uuid in collection according to JSON
	// in body
	editEntity(targetUuid uuid.UUID, body []byte) error

	// delete entity with targetUuid
	delEntity(targetUuid uuid.UUID) error
}

// type definition of a generic api entity
type entity interface{}

// user is an api entity
type user struct {
	Uuid       uuid.UUID
	FirstName  string
	SecondName string
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

func (uc *userCollection) createEntity(body []byte) error {
	var u user
	err := json.Unmarshal(body, (*userNew)(&u))
	if err != nil {
		return err
	}
	*uc = append(*uc, u)
	return nil
}

func (uc *userCollection) getEntity(targetUuid uuid.UUID) (entity, error) {
	var i int
	for i, _ = range *uc {
		if uuid.Equal((*uc)[i].Uuid, targetUuid) {
			return &(*uc)[i], nil
		}
	}
	return nil, errors.New("could not find user")
}

func (uc *userCollection) getCollection() (interface{}, error) {
	return uc, nil
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

func init() {
	users = []user{}
}

// checks log-in credentials
func verifyAccount(uname string, pword string) bool {
	if uname == "jcsharp" && pword == "pwd" {
		return true
	}
	return false
}

// returns two http.Handlers for dealing with REST API requests
// manipulating entities in entity collection 'ec'
// first return value is for dealing with requests ending in /<uuid> and
// handles api retrieval, edit, and deletion of single entity
// second return value is for dealing with requests dealing with whole collection,
// and handles creation of an entity in the collection, and retrieval
// of whole collection
func entityApiHandlerFactory(ec entityCollection) (http.Handler, http.Handler) {
	singularHandler := func(w http.ResponseWriter, r *http.Request) {
		userUuid, err := uuid.FromString(r.URL.Path)

		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "error decoding UUID", http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case "PUT":
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "error parsing request body: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// replace with entity.edit()
			err = ec.editEntity(userUuid, b)
			if err != nil {
				http.Error(w, "error editing entity: "+err.Error(), http.StatusInternalServerError)
				return
			}

			return
		case "DELETE":
			err = ec.delEntity(userUuid)
			if err != nil {
				http.Error(w, "error deleting entity: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "GET":
			var ej []byte
			u, err := ec.getEntity(userUuid)
			if err != nil {
				http.Error(w, "could not find user", http.StatusInternalServerError)
				return
			}
			ej, err = json.Marshal(u)
			if err != nil {
				http.Error(w, "error encoding JSON", http.StatusInternalServerError)
				return
			}

			fmt.Fprint(w, string(ej))
		default:
		}

	}

	pluralHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			var ej []byte
			c, err := ec.getCollection()
			if err != nil {
				http.Error(w, "error retrieving collection", http.StatusInternalServerError)
				return
			}

			ej, err = json.Marshal(c)
			if err != nil {
				http.Error(w, "error decoding JSON", http.StatusInternalServerError)
				return
			}

			fmt.Fprint(w, string(ej))
			return

		case "POST":
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "error parsing request body: "+err.Error(), http.StatusInternalServerError)
				return
			}
			err = ec.createEntity(b)
			if err != nil {
				http.Error(w, "error creating user: "+err.Error(), http.StatusInternalServerError)
				return
			}
		default:
		}
	}

	return http.HandlerFunc(singularHandler), http.HandlerFunc(pluralHandler)
}

func applySecurity(handler http.Handler) http.Handler {
	securityHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "OPTIONS" {
			var uname, pword, ok = r.BasicAuth()
			if !ok {
				w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
				w.Header().Add("WWW-Authenticate", "Basic realm=\"a\"")
				http.Error(w, "", http.StatusUnauthorized)
				return
			}
			if !verifyAccount(uname, pword) {
				http.Error(w, "incorrect uname/pword", http.StatusForbidden)
				return
			}
		}
		handler.ServeHTTP(w, r)
	}

	return http.HandlerFunc(securityHandler)
}

func applyCorsHeaders(handler http.Handler) http.Handler {
	corsHandler := func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "OPTIONS" {
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
			w.Header().Add("Access-Control-Allow-Headers", "Authorization")
			// TODO allow specification of the allowed methods
			w.Header().Add("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE")
			return
		} else if r.Method == "GET" || r.Method == "PUT" || r.Method == "POST" || r.Method == "DELETE" {
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
			handler.ServeHTTP(w, r)
		}
	}

	return http.HandlerFunc(corsHandler)
}

// takes a route to an entity collection and an entity collection
// and sets up handlers with defaultMux in net/http for entities of
// this type
func createApiRoute(path string, ec entityCollection) {
	sHandler, pHandler := entityApiHandlerFactory(ec)

    // apply security authorization
	sHandler = applySecurity(sHandler)
	pHandler = applySecurity(pHandler)

	// apply CORS headers
	sHandler = applyCorsHeaders(sHandler)
	pHandler = applyCorsHeaders(pHandler)

	http.Handle(path, pHandler)
	sPath := path + "/"
	http.Handle(sPath, http.StripPrefix(sPath, sHandler))
}

// basic part of api for validating a user
func verificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
	if r.Method == "OPTIONS" {
		w.Header().Add("Access-Control-Allow-Headers", "Authorization")
		w.Header().Add("Access-Control-Allow-Methods", "GET")
		return
	}

	var uname, pword, ok = r.BasicAuth()
	if !ok {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"a\"")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	if !verifyAccount(uname, pword) {
		http.Error(w, "incorrect uname/pword", http.StatusForbidden)
		return
	}
}

func main() {
	createApiRoute("/users", &users)

	http.HandleFunc("/verification", verificationHandler)

	http.ListenAndServe(":8080", nil)
}
