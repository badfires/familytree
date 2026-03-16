package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"family-tree/service"
)

func FamilyGraphHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	depth := 4
	if s := r.URL.Query().Get("depth"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			depth = v
		}
	}

	graph, err := service.BuildFamilyGraph(id, depth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(graph)
}