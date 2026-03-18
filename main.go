package main

import (
	"embed"
	"log"
	"net/http"

	"family-tree/database"
	"family-tree/handler"
	"family-tree/service"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func buildMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/person/create", handler.CreatePersonHandler)
	mux.HandleFunc("/person/get", handler.GetPersonHandler)
	mux.HandleFunc("/person/update", handler.UpdatePersonHandler)
	mux.HandleFunc("/person/search", handler.SearchPersonHandler)
	mux.HandleFunc("/person/suggest", handler.SearchSuggestHandler)

	mux.HandleFunc("/marriage/create", handler.CreateMarriageHandler)
	mux.HandleFunc("/marriage/get", handler.GetMarriageHandler)
	mux.HandleFunc("/marriage/update", handler.UpdateMarriageHandler)
	mux.HandleFunc("/marriage/add_child", handler.AddMarriageChildHandler)

	mux.HandleFunc("/adoption/create", handler.CreateAdoptionHandler)

	mux.HandleFunc("/person/export_template", handler.ExportPersonTemplateHandler)
	mux.HandleFunc("/person/import_csv", handler.ImportPersonCSVHandler)
	mux.HandleFunc("/person/min_id", handler.GetMinPersonIDHandler)

	mux.HandleFunc("/tree", handler.GetTreeHandler)
	mux.HandleFunc("/graph", handler.GraphHandler)
	mux.HandleFunc("/family_graph", handler.FamilyGraphHandler)
	mux.HandleFunc("/family_view", handler.FamilyViewHandler)

	return mux
}

func main() {
	database.InitDB()

	if err := service.EnsureSequencesInitialized(); err != nil {
		log.Fatal(err)
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "FamilyTree",
		Width:  1280,
		Height: 860,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: buildMux(),
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}