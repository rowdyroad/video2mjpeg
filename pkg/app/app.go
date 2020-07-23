package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"image/jpeg"
	"math/rand"
	"net/http"
	"video2mjpeg/pkg/caster"

	log "github.com/rowdyroad/go-simple-logger"
)


type Config struct {
	Listen string `yaml:"listen"`
}


type App struct {
	config Config
	server *http.Server
	caster *caster.Caster
}


func NewApp(config Config) *App {
	caster := caster.NewCaster()

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) {
		var req struct {
			Input string `form:"input"`
			FPS *int64 `form:"fps"`
			QScale *int64 `form:"qscale"`
			Scale *string `form:"scale"`
		}

		if err := c.BindQuery(&req); err != nil {
			return
		}
		stopChan := make(chan bool)
		imageChan, doneChan, err := caster.Cast(req.Input,req.FPS,req.QScale,req.Scale, stopChan)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError,err)
			return
		}
		boundary := rand.Int63()

		c.Header("Content-Type", fmt.Sprintf("multipart/x-mixed-replace;boundary=%v", boundary))
		c.Header("Pragma", "no-cache")
		defer func() { stopChan <- true }()

		for {
			select {
			case <-doneChan:
				log.Debug("Close http connection")
				return
			case image := <-imageChan:
				c.Writer.Write([]byte(fmt.Sprintf("\r\n--%v\r\n", boundary)))
				c.Writer.Write([]byte("Content-type: image/jpeg\r\n"))
				c.Writer.Write([]byte(fmt.Sprintf("Content-length: %d\r\n\r\n", 0)))
				jpeg.Encode(c.Writer, image, nil)
			}
		}
		close(stopChan)

	})
	server := &http.Server{
		Addr: config.Listen,
		Handler: router,
	}
	return &App{
		config,
		server,
		caster,
	}
}

func (a *App) Run() {
	a.server.ListenAndServe()
}

func (a *App) Close() {
	a.caster.Close()
	a.server.Shutdown(context.Background())
}
