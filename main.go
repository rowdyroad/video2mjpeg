package main

import (
	"fmt"
	"math/rand"
	"net/http"
)

func main() {

	caster := NewCaster()
	defer caster.Close()

	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		source := r.URL.Query().Get("source")
		dataChan, doneChan := caster.Cast(source, w.(http.CloseNotifier).CloseNotify())
		boundary := rand.Int63()
		w.Header().Add("Content-Type", fmt.Sprintf("multipart/x-mixed-replace;boundary=%v", boundary))
		w.Header().Add("Pragma", "no-cache")
		defer r.Body.Close()
		for {
			select {
			case <-doneChan:
				return

			case buf := <-dataChan:
				w.Write([]byte(fmt.Sprintf("\r\n--%v\r\n", boundary)))
				w.Write([]byte("Content-type: image/jpeg\r\n"))
				w.Write([]byte(fmt.Sprintf("Content-length: %d\r\n\r\n", len(buf))))
				w.Write(buf)
			}
		}
	})

	http.ListenAndServe(":8888", nil)
}
