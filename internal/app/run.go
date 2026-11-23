package app

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"io/fs"
	"kicau.desktop/cctv/internal/bindings"
)

func Run(assets fs.FS) error {
	service, e := bindings.New()
	if e != nil {
		return e
	}
	a := application.New(application.Options{Name: "Kicau Mania Desktop", Description: "CCTV, saving dan upload arena", Services: []application.Service{application.NewService(service)}, Assets: application.AssetOptions{Handler: application.AssetFileServerFS(assets)}})
	w := a.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Kicau Mania Desktop", Width: 1440, Height: 900, MinWidth: 960, MinHeight: 640, URL: "/",
		BackgroundColour: application.NewRGB(244, 247, 246),
		Windows:          application.WindowsWindow{Theme: application.Light},
	})
	bindings.Wire(service, a, w)
	return a.Run()
}
