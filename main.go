package main

import (
	"net/http"
)

func init() {
	users = []user{}
	threads = []thread{}
	http.Handle("/", rootApiHandler)
}

// checks log-in credentials
func verifyAccount(uname string, pword string) bool {
	if uname == "jcsharp" && pword == "pwd" {
		return true
	}
	return false
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
	createApiObject(&users)
	createApiObject(&threads)
	createApiObject(&messages)

	http.HandleFunc("/verification", verificationHandler)

	http.ListenAndServe(":8080", nil)
}
