package main

import (
	"github.com/TheAmgadX/linkit-abc-price-updates-cli-tool/gui"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

func main() {
	app := gui.NewApp()

	err := wails.Run(&options.App{
		Title:     "LinkIt — ABC Price Updater",
		Width:     620,
		Height:    720,
		MinWidth:  560,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets: gui.Assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
