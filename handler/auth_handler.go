package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

const AdminPasswordHeader = "X-Admin-Password"

type loginRequest struct {
	Password string `json:"password"`
}

func getAdminPassword() string {
	return strings.TrimSpace(os.Getenv("FAMILYTREE_ADMIN_PASSWORD"))
}

func extractAdminPassword(r *http.Request) string {
	pwd := strings.TrimSpace(r.Header.Get(AdminPasswordHeader))
	if pwd != "" {
		return pwd
	}

	pwd = strings.TrimSpace(r.URL.Query().Get("admin_password"))
	if pwd != "" {
		return pwd
	}

	_ = r.ParseForm()
	return strings.TrimSpace(r.FormValue("admin_password"))
}

func isAdminAuthorized(r *http.Request) bool {
	adminPwd := getAdminPassword()
	if adminPwd == "" {
		return false
	}
	return extractAdminPassword(r) == adminPwd
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdminAuthorized(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adminPwd := getAdminPassword()
	if adminPwd == "" {
		http.Error(w, "admin password not configured", http.StatusInternalServerError)
		return
	}

	if strings.TrimSpace(req.Password) != adminPwd {
		http.Error(w, "password error", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok": true,
	})
}

func AdminStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"configured": getAdminPassword() != "",
		"authorized": isAdminAuthorized(r),
	})
}