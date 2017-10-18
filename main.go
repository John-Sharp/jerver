package main

import (
	"context"
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
	http.Handle("/", rootApiHandler)
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

type key int

const userKey key = 0

// gets pointer to the user that created this
// request, set by a call to `applySecurity`
func getUserFromRequest(r *http.Request) *user {
	return r.Context().Value(userKey).(*user)
}

func applySecurity(handler http.Handler) http.Handler {
	securityHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			handler.ServeHTTP(w, r)
			return
		}

		var uname, pword, ok = r.BasicAuth()
		if !ok {
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
			w.Header().Add("WWW-Authenticate", "Basic realm=\"a\"")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		user, err := verifyAccount(uname, pword)
		if err != nil {
			http.Error(w, "incorrect uname/pword", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		handler.ServeHTTP(w, r.WithContext(ctx))
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
			w.Header().Add("Access-Control-Expose-Headers", "Location")
			handler.ServeHTTP(w, r)
		}
	}

	return http.HandlerFunc(corsHandler)
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
	createApiObject(&users)
	createApiObject(&threads)
	createApiObject(&messages)

	http.HandleFunc("/verification", verificationHandler)

	http.ListenAndServe(":8080", nil)
}
