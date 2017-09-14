package main

import (
    "net/http"
    "fmt"
    "encoding/json"
    "github.com/satori/go.uuid"
    "io/ioutil"
    "errors"
)

func readAndUnmarshal(r *http.Request, x interface{}) error {
    rb, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return err
    }
    return json.Unmarshal(rb, x)
}

type user struct {
    Uuid uuid.UUID
    FirstName string
    SecondName string
}

func (u * user) verifyAndParseNew(b []byte) error {
    var t struct{
        FirstName * string
        SecondName * string
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

func (u * user) verifyAndParseEdit(b []byte) error {
    bu := u.Uuid
    err := json.Unmarshal(b, &u)
    if err != nil {
        return err
    }
    u.Uuid = bu
    return nil
}

type userNew user
func (u * userNew) UnmarshalJSON(b []byte) error {
    return (* user)(u).verifyAndParseNew(b)
}

type userEdit user
func (u * userEdit) UnmarshalJSON(b []byte) error {
    return (* user)(u).verifyAndParseEdit(b)
}

var users []user

func findUserIndex(targetUuid uuid.UUID) (int, error) {
    var i int
    for i, _ = range users {
        if uuid.Equal(users[i].Uuid, targetUuid) {
            return i, nil
        }
    }
    return -1, errors.New("could not find user")
}

func init() {
    users = []user{}
}

func verifyUser (uname string, pword string) bool {
    if uname == "jcsharp" && pword == "pwd" {
        return true
    }
    return false
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "OPTIONS" {
        var uname, pword, ok = r.BasicAuth()
        if (!ok) {
            w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
            w.Header().Add("WWW-Authenticate", "Basic realm=\"a\"")
            http.Error(w, "", http.StatusUnauthorized)
            return
        }

        if !verifyUser(uname, pword) {
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

        var u user
        err := readAndUnmarshal(r, (*userNew)(&u))
        if err != nil {
            http.Error(w, "error parsing request body: " + err.Error(), http.StatusInternalServerError)
            return
        }

        users = append(users, u)
    default:
    }
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    userUuid, err := uuid.FromString(r.URL.Path)
    var i int

    if r.Method != "OPTIONS" {
        i, err = findUserIndex(userUuid)
        if err != nil {
            http.Error(w, "user not found", http.StatusInternalServerError)
            return
        }

        var uname, pword, ok = r.BasicAuth()
        if (!ok) {
            w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
            w.Header().Add("WWW-Authenticate", "Basic realm=\"a\"")
            http.Error(w, "", http.StatusUnauthorized)
            return
        }
        if !verifyUser(uname, pword) {
            http.Error(w, "incorrect uname/pword", http.StatusForbidden)
            return
        }

    }

    if err != nil {
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

        err = readAndUnmarshal(r, (*userEdit)(&users[i]))

        if err != nil {
            http.Error(w, "error parsing request body: " + err.Error(), http.StatusInternalServerError)
            return
        }

        return
    case "DELETE":
        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")
        users = append(users[:i], users[i+1:]...)
    case "GET":
        w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8090")

        var ej []byte
        ej , err := json.Marshal(users[i])
        if err != nil {
            http.Error(w, "error encoding JSON", http.StatusInternalServerError)
            return
        }

        fmt.Fprint(w, string(ej))
    default:
    }
}

func main() {
    http.HandleFunc("/users", usersHandler)
    http.Handle("/users/", http.StripPrefix("/users/", http.HandlerFunc(userHandler)))

    http.ListenAndServe(":8080", nil)
}
