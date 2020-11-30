package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var c *cache

func doGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	val, err := c.Get(key)
	if err != nil {
		if err == ErrNotFound {
			http.Error(w, "Key Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Unknown error", http.StatusInternalServerError)
		}
		return
	}
	w.Write(val.([]byte))
}

func doPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request body", http.StatusBadRequest)
		return
	}
	c.Put(key, data)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c = NewCache(ctx)

	r := mux.NewRouter()
	r.HandleFunc("/{key}", doGet).Methods("GET")
	r.HandleFunc("/{key}", doPost).Methods("POST")
	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
