package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
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
		log.Println("Received http request. Command  is %v", command)
		stopChan := make(chan bool)
		headerChan, dataChan, doneChan, err := caster.Cast(command, stopChan)
		if err != nil {
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
				return
			case <-doneChan:
				log.Println("Close http connection")
				return
			case buf := <-dataChan:
				w.Write(buf)
			case <-headerChan:
				w.Write([]byte(fmt.Sprintf("\r\n--%v\r\n", boundary)))
				w.Write([]byte("Content-type: image/jpeg\r\n"))
				w.Write([]byte(fmt.Sprintf("Content-length: %d\r\n\r\n", 0)))
			}
		}

	})

	http.ListenAndServe(listen, nil)
}
