package main

import (
    "net/http"
    "fmt"
    //"log"
    "encoding/json"
    // "bytes"
    // "net/url"
)

type person struct {
    FirstName string
    SecondName string
}

var employees []person

func barHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var ej []byte
        ej , err := json.Marshal(employees)
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

        employees = append(employees, person{r.PostForm["firstName"][0], r.PostForm["secondName"][0]})
    default:
    }
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "got foo handler")
}

func main() {
    http.HandleFunc("/bar", barHandler)
    http.HandleFunc("/foo", fooHandler)

    http.ListenAndServe(":8080", nil)
}
