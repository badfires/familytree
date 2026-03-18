package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"family-tree/model"
	"family-tree/service"
)

type createAdoptionRequest struct {
	PersonID   string `json:"person_id"`
	ToFatherID string `json:"to_father_id"`
	ToMotherID string `json:"to_mother_id"`
	Note       string `json:"note"`
}

func CreateAdoptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createAdoptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	a := model.Adoption{
		PersonID:   strings.TrimSpace(req.PersonID),
		ToFatherID: strings.TrimSpace(req.ToFatherID),
		ToMotherID: strings.TrimSpace(req.ToMotherID),
		Note:       strings.TrimSpace(req.Note),
	}

	created, err := service.CreateAdoption(a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(created)
}

func GetAdoptionHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	personID := strings.TrimSpace(r.URL.Query().Get("person_id"))

	var (
		a   *model.Adoption
		err error
	)

	switch {
	case id != "":
		a, err = service.GetAdoptionByID(id)
	case personID != "":
		a, err = service.GetAdoption(personID)
	default:
		http.Error(w, "missing id or person_id", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(a)
}

func UpdateAdoptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var a model.Adoption
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	a.ID = strings.TrimSpace(a.ID)
	a.PersonID = strings.TrimSpace(a.PersonID)
	a.FromFatherID = strings.TrimSpace(a.FromFatherID)
	a.FromMotherID = strings.TrimSpace(a.FromMotherID)
	a.ToFatherID = strings.TrimSpace(a.ToFatherID)
	a.ToMotherID = strings.TrimSpace(a.ToMotherID)
	a.Note = strings.TrimSpace(a.Note)

	if err := service.UpdateAdoption(a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok": true,
		"id": a.ID,
	})
}