package main

import (
	"io"
	"sync"

	log "github.com/rowdyroad/go-simple-logger"
)

//Source video
type Source struct {
	sync.Mutex
	Streams []*Stream
	Pipe    io.ReadCloser
	Stop    bool
}

//Close source pipe and all streams
func (s *Source) Close() {
	log.Debug("Closing source")
	s.Stop = true
	for index, stream := range s.Streams {
		log.Debug("Closing stream", index)
		stream.DoneChan <- true
		log.Debug("Stream", index, "closed")
	}
	s.Pipe.Close()
	log.Debug("Pipe is closed")
}
