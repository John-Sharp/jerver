package main

import (
    "net/http"
    "fmt"
    "encoding/json"
    "github.com/satori/go.uuid"
)

type user struct {
    Uuid uuid.UUID
    FirstName string
    SecondName string
}

var users []user

func init() {
    users = []user{}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var ej []byte
        ej , err := json.Marshal(users)
        if err != nil {
            http.Error(w, "error decoding JSON", http.StatusInternalServerError)
            return
        }

        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
        fmt.Fprint(w, string(ej))
        return

    case "POST":
        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
        firstName := r.PostFormValue("firstName")
        secondName := r.PostFormValue("secondName")

        if firstName == "" || secondName == "" {
            http.Error(w, "error parsing form", http.StatusInternalServerError)
            return
        }

        users = append(users, user{
            uuid.NewV4(),
            r.PostForm["firstName"][0],
            r.PostForm["secondName"][0]})
    default:
    }
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    userUuid, err := uuid.FromString(r.URL.Path)
    if err != nil {
        http.Error(w, "error decoding UUID", http.StatusInternalServerError)
        return
    }

    switch r.Method {
    case "OPTIONS":
        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
        w.Header().Add("Access-Control-Allow-Methods", "GET, PUT, DELETE")
        return
    case "PUT":
        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
        found := false
        for i, u := range users {
            if uuid.Equal(u.Uuid, userUuid) {
                found = true

                if r.PostFormValue("firstName") != "" {
                    users[i].FirstName = r.PostFormValue("firstName")
                }
                if r.PostFormValue("secondName") != "" {
                    users[i].SecondName = r.PostFormValue("secondName")
                }
                break
            }
        }

        if !found {
            http.Error(w, "user not found", http.StatusInternalServerError)
            return
        }

        return
    case "DELETE":
        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
        found := false
        var i int
        var u user
        for i, u = range users {
            if uuid.Equal(u.Uuid, userUuid) {
                found = true
            }
            break
        }

        if !found {
            http.Error(w, "user not found", http.StatusInternalServerError)
            return
        }

        users = append(users[:i], users[i+1:]...)
    case "GET":
        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
        found := false

        var i int
        var u user
        for i, u = range users {
            if uuid.Equal(u.Uuid, userUuid) {
                found = true
            }
            break
        }

        if !found {
            http.Error(w, "user not found", http.StatusInternalServerError)
            return
        }

        var ej []byte
        ej , err := json.Marshal(users[i])
        if err != nil {
            http.Error(w, "error decoding JSON", http.StatusInternalServerError)
            return
        }

        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
        fmt.Fprint(w, string(ej))
    default:
    }
}

func main() {
    http.HandleFunc("/users", usersHandler)
    http.Handle("/users/", http.StripPrefix("/users/", http.HandlerFunc(userHandler)))

    http.ListenAndServe(":8080", nil)
}
