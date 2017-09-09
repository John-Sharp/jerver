package main

import (
    "net/http"
    "fmt"
    //"log"
    "encoding/json"
    // "bytes"
    // "net/url"
)

type user struct {
    FirstName string
    SecondName string
    Password string
}

var users []user

func userHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var ej []byte
        ej , err := json.Marshal(users)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        fmt.Fprint(w, string(ej))
        return

    case "POST":
        err := r.ParseForm()
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        if val, ok := r.PostForm["firstName"]; ok {
            if len(val) != 1 {
                w.WriteHeader(http.StatusBadRequest)
                return
            }
        }
        if val, ok := r.PostForm["secondName"]; ok {
            if len(val) != 1 {
                w.WriteHeader(http.StatusBadRequest)
                return
            }
        }
        if val, ok := r.PostForm["password"]; ok {
            if len(val) != 1 {
                w.WriteHeader(http.StatusBadRequest)
                return
            }
        }

        users = append(users, user{
            r.PostForm["firstName"][0],
            r.PostForm["secondName"][0],
            r.PostForm["password"][0]})
    default:
    }
}

func main() {
    http.HandleFunc("/users", userHandler)

    http.ListenAndServe(":8080", nil)
}
