package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"techbrewers.com/usr/repos/entitycoll"
)

var db *sql.DB

func init() {
	var err error

	db, err = sql.Open("sqlite3", "./jerver.db")
	if err != nil {
		log.Fatal(err)
	}

	users.prepareStmts()
	threads.prepareStmts()
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
	_, err := authorizeUser(uname, pword)
	if err != nil {
		http.Error(w, "incorrect uname/pword", http.StatusForbidden)
		return
	}
}

func main() {
	entitycoll.Configure(entitycoll.Configuration{ApiRoot: "/", AccessControlAllowOrigin: "http://localhost:8090", RequestorAuthFn: authorizeUser})
	entitycoll.CreateApiObject(&users)
	entitycoll.CreateApiObject(&threads)
	entitycoll.CreateApiObject(&messages)

	http.HandleFunc("/verification", verificationHandler)

	http.ListenAndServe(":8080", nil)
}
