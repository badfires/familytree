package handler

import (
	"net/http"

	"family-tree/service"
)

func AddMarriageChildHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	marriageID := r.URL.Query().Get("marriage_id")
	childID := r.URL.Query().Get("child_id")
	if marriageID == "" || childID == "" {
		http.Error(w, "missing marriage_id or child_id", http.StatusBadRequest)
		return
	}

	if err := service.AddChildToMarriage(marriageID, childID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("ok"))
}