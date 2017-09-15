package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
)

type entityCollection interface {
	createEntity(body []byte) error
	getEntity(targetUuid uuid.UUID) (entity, error)
	getCollection() (interface{}, error)
	editEntity(targetUuid uuid.UUID, body []byte) error
	delEntity(targetUuid uuid.UUID) error
}
type entity interface{}

type user struct {
	Uuid       uuid.UUID
	FirstName  string
	SecondName string
}

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
func (u *user) verifyAndParseEdit(b []byte) error {
	bu := u.Uuid
	err := json.Unmarshal(b, &u)
	if err != nil {
		return err
	}
	u.Uuid = bu
	return nil
}

type userNew user

func (u *userNew) UnmarshalJSON(b []byte) error {
	return (*user)(u).verifyAndParseNew(b)
}

type userEdit user

func (u *userEdit) UnmarshalJSON(b []byte) error {
	return (*user)(u).verifyAndParseEdit(b)
}

type userCollection []user

var users userCollection

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

func verifyAccount(uname string, pword string) bool {
	if uname == "jcsharp" && pword == "pwd" {
		return true
	}
	return false
}

func entityApiHandlerFactory(ec entityCollection) (http.Handler, http.Handler) {
	singularHandler := func(w http.ResponseWriter, r *http.Request) {
		userUuid, err := uuid.FromString(r.URL.Path)

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

		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "error decoding UUID", http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case "OPTIONS":
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
			w.Header().Add("Access-Control-Allow-Headers", "Authorization")
			w.Header().Add("Access-Control-Allow-Methods", "GET, PUT, DELETE")
			return
		case "PUT":
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")

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
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")

			err = ec.delEntity(userUuid)
			if err != nil {
				http.Error(w, "error deleting entity: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "GET":
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")

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

		switch r.Method {
		case "OPTIONS":
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
			w.Header().Add("Access-Control-Allow-Headers", "Authorization")
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST")
			return
		case "GET":
			var ej []byte
			ej, err := json.Marshal(users)
			if err != nil {
				http.Error(w, "error decoding JSON", http.StatusInternalServerError)
				return
			}

			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
			fmt.Fprint(w, string(ej))
			return

		case "POST":
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")

			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "error parsing request body: "+err.Error(), http.StatusInternalServerError)
				return
			}
			err = users.createEntity(b)
			if err != nil {
				http.Error(w, "error creating user: "+err.Error(), http.StatusInternalServerError)
				return
			}
		default:
		}
	}

	return http.HandlerFunc(singularHandler), http.HandlerFunc(pluralHandler)
}

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
	sHandler, pHandler := entityApiHandlerFactory(&users)

	http.Handle("/users", pHandler)
	http.Handle("/users/", http.StripPrefix("/users/", sHandler))

	http.HandleFunc("/verification", verificationHandler)

	http.ListenAndServe(":8080", nil)
}
