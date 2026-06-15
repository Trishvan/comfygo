package main

import (
	"context"
	"embed"
	"io/fs"
	"log"

	"github.com/comfygo/backend/orchestrator"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var frontendDist embed.FS

func main() {
	mgr := orchestrator.NewManager()

	subFS, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	err = wails.Run(&options.App{
		Title:     "ComfyGo",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 600,
		MaxWidth:  9999,
		MaxHeight: 9999,
		AssetServer: &assetserver.Options{
			Assets:  subFS,
			Handler: mgr.AssetHandler,
		},
		OnStartup: func(ctx context.Context) {
			mgr.Start(ctx)
		},
		OnDomReady: func(ctx context.Context) {
			wailsRuntime.WindowSetMaxSize(ctx, 0, 0)
		},
		Bind: []interface{}{
			mgr,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
