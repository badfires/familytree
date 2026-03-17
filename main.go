package main

import (
	"log"
	"net/http"

	"family-tree/database"
	"family-tree/handler"
	"family-tree/service"
)

func main() {
	database.InitDB()

	if err := service.EnsureSequencesInitialized(); err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/person/create", handler.CreatePersonHandler)
	http.HandleFunc("/person/get", handler.GetPersonHandler)
	http.HandleFunc("/person/update", handler.UpdatePersonHandler)
	http.HandleFunc("/person/search", handler.SearchPersonHandler)
	http.HandleFunc("/person/suggest", handler.SearchSuggestHandler)

	http.HandleFunc("/marriage/create", handler.CreateMarriageHandler)
	http.HandleFunc("/marriage/get", handler.GetMarriageHandler)
	http.HandleFunc("/marriage/update", handler.UpdateMarriageHandler)
	http.HandleFunc("/marriage/add_child", handler.AddMarriageChildHandler)

	http.HandleFunc("/adoption/create", handler.CreateAdoptionHandler)
		http.HandleFunc("/person/export_template", handler.ExportPersonTemplateHandler)
	http.HandleFunc("/person/import_csv", handler.ImportPersonCSVHandler)

	http.HandleFunc("/tree", handler.GetTreeHandler)
	http.HandleFunc("/graph", handler.GraphHandler)
	http.HandleFunc("/family_graph", handler.FamilyGraphHandler)
	http.HandleFunc("/family_view", handler.FamilyViewHandler)
http.HandleFunc("/person/min_id", handler.GetMinPersonIDHandler)
	log.Println("server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}