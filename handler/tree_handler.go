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

	_ = json.NewEncoder(w).Encode(tree)
}