package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"family-tree/service"
)

func GraphHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	depth := 3
	if s := r.URL.Query().Get("depth"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			depth = v
		}
	}

	tree, err := service.BuildGraph(id, depth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(tree)
}