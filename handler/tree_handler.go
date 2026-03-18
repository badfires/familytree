package handler

import (
	"encoding/json"
	"net/http"

	"family-tree/service"
)

func GetTreeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	tree, err := service.GetTree(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(tree)
}