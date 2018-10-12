package main

import (
	"io"
	"sync"

	log "github.com/rowdyroad/go-simple-logger"
)

type Source struct {
	sync.Mutex
	Streams []Stream
	Pipe    io.ReadCloser
	Stop    bool
}

func (s *Source) Close(broadcast bool) {
	log.Debug("Closing source / broadcasted:", broadcast)
	s.Lock()
	defer s.Unlock()
	s.Stop = true
	s.Pipe.Close()
	log.Debug("Pipe is closed")
	if broadcast {
		log.Debug("Broadcasting...")
		for index, stream := range s.Streams {
			log.Debug("Closing stream", index)
			stream.DoneChan <- true
			log.Debug("Stream", index, "closed")
		}
	}
}
