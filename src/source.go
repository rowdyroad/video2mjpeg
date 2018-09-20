package main

import (
	"io"
	"sync"
)

type Source struct {
	sync.Mutex
	Streams []Stream
	Pipe    io.ReadCloser
	Stop    bool
}

func (s *Source) Close(broadcast bool) {
	s.Lock()
	defer s.Unlock()
	s.Stop = true
	s.Pipe.Close()
	if broadcast {
		for _, stream := range s.Streams {
			stream.DoneChan <- true
		}
	}
}
