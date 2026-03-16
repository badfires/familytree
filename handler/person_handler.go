package handler

import (
	"encoding/json"
	"net/http"

	"family-tree/model"
	"family-tree/service"
)

func CreatePersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p model.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := service.CreatePerson(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("ok"))
}

func GetPersonHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	p, err := service.GetPerson(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(p)
}

func UpdatePersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p model.Person
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := service.UpdatePerson(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("ok"))
}

func SearchPersonHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	list, err := service.SearchPerson(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(list)
}

func SearchSuggestHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	list, err := service.SearchPerson(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(list)
}