package main

import (
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"strings"
)

// ServeMux for storing direct paths to entities
// the `rootApiHandler` will process the
// url it receives and look for entities to call
// the handler of
var entityServeMux http.ServeMux

// interface for generic collection of api entities
type entityCollection interface {

	// gets the name of the URL component referring to this entity
	getRestName() string

	// get a pointer to an entityCollection that is the parent
	// of this entity collection (i.e. that's path in the API
	// preceeds a mention of this entity)
	getParentCollection() entityCollection

	// given a []byte containing JSON, and the url path of
	// the REST request should create an entity and
	// add it to the collection
	// returns the REST path to the newly created entity
	createEntity(parentEntityUuids map[string]uuid.UUID, body []byte) (string, error)

	// given a Uuid should find entity in collection and return
	getEntity(targetUuid uuid.UUID) (entity, error)

	// return whole collection located at urlPath
	// (in future maybe include filters
	// in argument and return subset of collection)
	getCollection(parentEntityUuids map[string]uuid.UUID) (interface{}, error)

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
		pathComponents := strings.Split(r.URL.Path, "/")[1:]
		entityUuid, err := uuid.FromString(pathComponents[len(pathComponents)-1])

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
			err = ec.editEntity(entityUuid, b)
			if err != nil {
				http.Error(w, "error editing entity: "+err.Error(), http.StatusInternalServerError)
				return
			}

			return
		case "DELETE":
			err = ec.delEntity(entityUuid)
			if err != nil {
				http.Error(w, "error deleting entity: "+err.Error(), http.StatusInternalServerError)
				return
			}
		case "GET":
			var ej []byte
			u, err := ec.getEntity(entityUuid)
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
		pathComponents := strings.Split(r.URL.Path, "/")[1:]

		if len(pathComponents)%2 != 1 {
			fmt.Println("collection entity URL should have an even number of components (entity name and UUID for each parent entity and name for entity)")
			http.Error(w, "error parsing URL", http.StatusInternalServerError)
			return
		}

		var err error
		parentEntityUuids := make(map[string]uuid.UUID)
		for i := 0; i < len(pathComponents)-1; i += 2 {
			parentEntityUuids[pathComponents[i]], err = uuid.FromString(pathComponents[i+1])

			if err != nil {
				fmt.Println("error decoding UUID of path component: ", pathComponents[i], ": ", err.Error())
				http.Error(w, "error parsing URL", http.StatusInternalServerError)
				return
			}
		}

		switch r.Method {
		case "GET":
			var ej []byte
			c, err := ec.getCollection(parentEntityUuids)
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
			entityPath, err := ec.createEntity(parentEntityUuids, b)
			if err != nil {
				http.Error(w, "error creating entity: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Location", entityPath)
			w.WriteHeader(http.StatusCreated)
		default:
		}
	}

	return http.HandlerFunc(singularHandler), http.HandlerFunc(pluralHandler)
}

// handles all requests to the api root, processes the requested URL
// to see what entity the request deals with and gets that handler to
// serve the request
var rootApiHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	pathBu := r.URL.Path

	// split url into components
	pathComponents := strings.Split(r.URL.Path, "/")

	// first hypothesis: request for collection of entities, where
	// final component of path is entity name
	entityName := pathComponents[len(pathComponents)-1]
	// see if there is a handler for this
	r.URL.Path = "/" + entityName
	h, pattern := entityServeMux.Handler(r)
	if pattern != "" {
		r.URL.Path = pathBu
		h.ServeHTTP(w, r)
		return
	}

	// second hypothesis: request for single entity, where
	// final component is entity id and penultimate component
	// is entity name
	entityName = pathComponents[len(pathComponents)-2]
	r.URL.Path = "/" + entityName + "/"
	h, pattern = entityServeMux.Handler(r)
	if pattern != "" {
		r.URL.Path = pathBu
		h.ServeHTTP(w, r)
		return
	}

	// no patterns found. Can just call ServeHTTP
	// on handler returned by failed search, since
	// it will be a not found handler
	r.URL.Path = pathBu
	h.ServeHTTP(w, r)
}

// takes a route to an entity collection and an entity collection
// and sets up handlers with defaultMux in net/http for entities of
// this type
func createApiRoute(path string, ec entityCollection) {
	sHandler, pHandler := entityApiHandlerFactory(ec)

	// apply security authorization
	sHandler = applySecurity(sHandler)
	pHandler = applySecurity(pHandler)

	// apply CORS headers
	sHandler = applyCorsHeaders(sHandler)
	pHandler = applyCorsHeaders(pHandler)

	entityServeMux.Handle(path, pHandler)
	sPath := path + "/"
	entityServeMux.Handle(sPath, sHandler)
}
