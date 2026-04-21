package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"family-tree/service"
)

type addMarriageChildRequest struct {
	MarriageID string `json:"marriage_id"`
	ChildID    string `json:"child_id"`
}

func AddMarriageChildHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	marriageID := strings.TrimSpace(r.URL.Query().Get("marriage_id"))
	childID := strings.TrimSpace(r.URL.Query().Get("child_id"))

	if marriageID == "" || childID == "" {
		var req addMarriageChildRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if marriageID == "" {
				marriageID = strings.TrimSpace(req.MarriageID)
			}
			if childID == "" {
				childID = strings.TrimSpace(req.ChildID)
			}
		}
	}

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
