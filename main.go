package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"vectordb-1/backend/service"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	connSvc := service.NewConnectionService()
	vdbSvc := service.NewVectorDBService()

	app := application.New(application.Options{
		Name:        "Tencent Cloud VectorDB Client",
		Description: "GUI Client for Tencent Cloud VectorDB",
		Services: []application.Service{
			application.NewService(connSvc),
			application.NewService(vdbSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Tencent Cloud VectorDB Client",
		Width:     1280,
		Height:    800,
		MinWidth:  960,
		MinHeight: 600,
		URL:       "/",
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
