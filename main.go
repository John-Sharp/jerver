package main

import (
	"net/http"

	"github.com/john-sharp/entitycoll"
)

func init() {
	var u user
	u.popNew("John", "Sharp", "jcsharp", "pwd")
	users = []user{
		u,
	}

	threads = []thread{}
	http.Handle("/", entitycoll.RootApiHandler)
}

// TODO deprecate
// checks log-in credentials SOON TO BE DEPRECATED
func verifyAccount(uname string, pwd string) (*user, error) {
	e, err := users.verifyUser(uname, pwd)
	u := (e).(*user)
	return u, err
}

func authorizeUser(uname, pwd string) (entitycoll.Entity, error) {
	return users.verifyUser(uname, pwd)
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
	_, err := verifyAccount(uname, pword)
	if err != nil {
		http.Error(w, "incorrect uname/pword", http.StatusForbidden)
		return
	}
}

func main() {
	entitycoll.SetRequestorAuthFn(authorizeUser)
	entitycoll.CreateApiObject(&users)
	entitycoll.CreateApiObject(&threads)
	entitycoll.CreateApiObject(&messages)

	http.HandleFunc("/verification", verificationHandler)

	http.ListenAndServe(":8080", nil)
}
