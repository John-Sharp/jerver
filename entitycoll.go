package main

import (
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
)

// interface for generic collection of api entities
type entityCollection interface {

	// given a []byte containing JSON, and the url path of
	// the REST request should create an entity and
	// add it to the collection
	createEntity(body []byte, urlPath string) error

	// given a Uuid should find entity in collection and return
	getEntity(targetUuid uuid.UUID) (entity, error)

	// return whole collection located at urlPath
	// (in future maybe include filters
	// in argument and return subset of collection)
	getCollection(urlPath string) (interface{}, error)

	// edit entity with Uuid in collection according to JSON
	// in body
	editEntity(targetUuid uuid.UUID, body []byte) error

	// delete entity with targetUuid
	delEntity(targetUuid uuid.UUID) error
}

// type definition of a generic api entity
type entity interface{}

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
				http.Error(w, "could not find entity", http.StatusInternalServerError)
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
			c, err := ec.getCollection(r.URL.Path)
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
			err = ec.createEntity(b, r.URL.Path)
			if err != nil {
				http.Error(w, "error creating entity: "+err.Error(), http.StatusInternalServerError)
				return
			}
		default:
		}
	}

	return http.HandlerFunc(singularHandler), http.HandlerFunc(pluralHandler)
}
