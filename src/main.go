package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"math/rand"
	"net/http"

	log "github.com/rowdyroad/go-simple-logger"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lshortlevel | log.Lcolor)
	var listen string
	flag.StringVar(&listen, "listen", ":80", "Listen address and port")
	flag.Parse()

	caster := NewCaster()
	defer caster.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		query := r.URL.Query()
		command := map[string]string{}
		for name, values := range query {
			command[name] = values[0]
		}
		if len(command) == 0 {
			w.WriteHeader(500)
			return
		}
		log.Debugf("Received http request. Command  is %v", command)
		stopChan := make(chan bool)
		imageChan, doneChan, err := caster.Cast(command, stopChan)
		if err != nil {
			log.Error("Cast create error:", err)
			w.WriteHeader(500)
			return
		}
		boundary := rand.Int63()
		w.Header().Add("Content-Type", fmt.Sprintf("multipart/x-mixed-replace;boundary=%v", boundary))
		w.Header().Add("Pragma", "no-cache")
		defer func() { stopChan <- true }()

		for {
			select {
			case <-w.(http.CloseNotifier).CloseNotify():
				log.Debug("HTTP close event")
				return
			case <-doneChan:
				log.Debug("Close http connection")
				return
			case image := <-imageChan:
				w.Write([]byte(fmt.Sprintf("\r\n--%v\r\n", boundary)))
				w.Write([]byte("Content-type: image/jpeg\r\n"))
				w.Write([]byte(fmt.Sprintf("Content-length: %d\r\n\r\n", 0)))
				jpeg.Encode(w, image, nil)
			}
		}
		close(stopChan)

	})

	http.ListenAndServe(listen, nil)
}
