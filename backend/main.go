package backend

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type badRequest struct{ error }
type notFound struct{ error }

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/_ah/api/echo", errorHandler(echoHandler)).Methods("GET", "POST")
	r.HandleFunc("/_ah/api/listings", errorHandler(listingsHandler)).Methods("GET", "POST")
	r.HandleFunc("/_ah/api/listing/{id}", errorHandler(listingHandler)).Methods("GET", "DELETE")
	http.Handle("/", r)
}

func echoHandler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	e := EchoModel{
		CurrentDate: time.Now(),
		Description: "Our Places API - echo",
		Name:        "Our Places",
	}

	if err := json.NewEncoder(w).Encode(e); err != nil {
		return err
	}

	return nil
}

// listingsHandler provides: GET: /listings, POST: /listings
func listingsHandler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		return getListings(w, r)
	case "POST":
		return addListing(w, r)
	}
	return nil
}

func listingHandler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		return getListing(w, r)
	case "DELETE":
		return deleteListing(w, r)
	}
	return nil
}

func getListings(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	listings, cursor, err := GetAllListings(c, 100, "")

	if err != nil {
		log.Errorf(c, "Error getting listings: %v", err)
	}

	ls := new(Listings)
	ls.Listings = listings
	ls.Next = cursor

	if err := json.NewEncoder(w).Encode(ls); err != nil {
		return err
	}

	return nil
}

func getListing(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		log.Errorf(c, "Invalid ID: %v", err)
	}
	l, err := GetListingByKey(c, id)

	if err != nil {
		return err
	}

	if err := json.NewEncoder(w).Encode(l); err != nil {
		log.Errorf(c, "Could not encode listing: %v", err)
		return err
	}

	return nil
}

func addListing(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var l Listing
	if err = json.Unmarshal(body, &l); err != nil {
		return err
	}

	key, err := AddListing(c, &l)
	if nil != err {
		return err
	}

	l.ID = key.IntID()
	if err := json.NewEncoder(w).Encode(l); err != nil {
		return err
	}
	return nil
}

func deleteListing(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Errorf(c, "Invalid ID: %v", err)
	}
	l, err := GetListingByKey(c, id)

	if err != nil {
		return err
	}

	err = DeleteListing(c, l.Key(c))
	if nil != err {
		return err
	}
	response := &ResultResponse{
		IsSuccessful: true,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return err
	}
	return nil
}

func errorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}
		switch err.(type) {
		case badRequest:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case notFound:
			http.Error(w, "task not found", http.StatusNotFound)
		default:
			//log.Errorf(c, "Unexpected error: %v", err)
			http.Error(w, "oops", http.StatusInternalServerError)
		}
	}
}
