package main

import (
	"context"
	"log"

	"github.com/comfygo/backend/orchestrator"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	mgr := orchestrator.NewManager()

	app := wails.Application{
		Options: &options.App{
			Title:  "ComfyGo",
			Width:  1280,
			Height: 800,
			AssetServer: &assetserver.Options{
				Handler: mgr.AssetHandler,
			},
			OnStartup: func(ctx context.Context) {
				mgr.Start(ctx)
			},
			Bind: []interface{}{
				mgr,
			},
		},
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
