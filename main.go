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

	// 鉴权
	mux.HandleFunc("/auth/login", handler.AdminLoginHandler)
	mux.HandleFunc("/auth/status", handler.AdminStatusHandler)

	// 人物
	mux.HandleFunc("/person/get", handler.GetPersonHandler)
	mux.HandleFunc("/person/search", handler.SearchPersonHandler)
	mux.HandleFunc("/person/suggest", handler.SearchSuggestHandler)
	mux.HandleFunc("/person/export_template", handler.ExportPersonTemplateHandler)
	mux.HandleFunc("/person/min_id", handler.GetMinPersonIDHandler)

	// 树图 / 详情
	mux.HandleFunc("/tree", handler.GetTreeHandler)
	mux.HandleFunc("/graph", handler.GraphHandler)
	mux.HandleFunc("/family_graph", handler.FamilyGraphHandler)
	mux.HandleFunc("/family_view", handler.FamilyViewHandler)

	// 婚姻读取
	mux.HandleFunc("/marriage/get", handler.GetMarriageHandler)

	// 过继读取
	mux.HandleFunc("/adoption/get", handler.GetAdoptionHandler)

	// 写接口：管理员权限
	mux.HandleFunc("/person/create", handler.RequireAdmin(handler.CreatePersonHandler))
	mux.HandleFunc("/person/update", handler.RequireAdmin(handler.UpdatePersonHandler))
	mux.HandleFunc("/person/import_csv", handler.RequireAdmin(handler.ImportPersonCSVHandler))

	mux.HandleFunc("/marriage/create", handler.RequireAdmin(handler.CreateMarriageHandler))
	mux.HandleFunc("/marriage/update", handler.RequireAdmin(handler.UpdateMarriageHandler))
	mux.HandleFunc("/marriage/add_child", handler.RequireAdmin(handler.AddMarriageChildHandler))

	mux.HandleFunc("/adoption/create", handler.RequireAdmin(handler.CreateAdoptionHandler))
	mux.HandleFunc("/adoption/update", handler.RequireAdmin(handler.UpdateAdoptionHandler))

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