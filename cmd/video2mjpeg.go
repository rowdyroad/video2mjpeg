package main

import (
	yconfig "github.com/rowdyroad/go-yaml-config"
	"github.com/rowdyroad/video2mjpeg/pkg/app"
)

func main() {
	var config app.Config
	yconfig.LoadConfig(&config, "video2mjpeg.yaml", nil)
	app := app.NewApp(config)
	defer app.Close()
	app.Run()
}


