package main

import (
	"embed"
	"kicau.desktop/cctv/internal/app"
	"log"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if e := app.Run(assets); e != nil {
		log.Fatal(e)
	}
}
