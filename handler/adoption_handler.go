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