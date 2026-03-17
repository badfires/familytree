package handler

import (
	"encoding/json"
	"net/http"

	"family-tree/service"
)

func ExportPersonTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := service.BuildPersonTemplateCSV()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="people_template.csv"`)
	w.Write(data)
}

func ImportPersonCSVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "缺少上传文件字段 file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	result, err := service.ImportPeopleCSV(file)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   err.Error(),
			"row":     resultRow(result),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func resultRow(result *service.PersonCSVImportResult) int {
	if result == nil {
		return 0
	}
	return result.Row
}