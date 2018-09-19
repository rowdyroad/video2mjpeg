package main

import (
	"io"
	"log"
	"os/exec"
	"sync"
)

type stream struct {
	DataChan chan []byte
	DoneChan chan bool
}

//NewCaster creates Caster instance
func NewCaster() *Caster {
	return &Caster{sync.Mutex{}, map[string]*[]stream{}}
}

//Caster converting rtsp(or any) stream to mjpeg (with ffmjpeg support)
type Caster struct {
	sync.Mutex
	streams map[string]*[]stream
}

func (c *Caster) Close() {
	c.Lock()
	defer c.Unlock()

	for _, streams := range c.streams {
		for _, stream := range *streams {
			stream.DoneChan <- true
		}
	}
}

//Cast main functionality for convert rtsp(or any) to mjpeg
func (c *Caster) Cast(source string, stopChan <-chan bool) (chan []byte, chan bool) {
	c.Lock()
	defer c.Unlock()

	log.Println("Castring for", source)

	streams, exists := c.streams[source]
	if !exists {
		c.streams[source] = &[]stream{}
		streams = c.streams[source]
	}
	current := stream{make(chan []byte), make(chan bool)}
	*streams = append(*streams, current)

	stop := false

	doneBroadcast := func() {
		log.Println("Done broadcasting")
		c.Lock()
		defer c.Unlock()
		for _, cts := range *streams {
			cts.DoneChan <- true
		}
	}

	go func() {
		<-stopChan
		log.Println("Client gone")
		c.Lock()
		defer c.Unlock()
		for index, cts := range *streams {
			if cts == current {
				log.Println("Removing client stream record.")
				*streams = append((*streams)[:index], (*streams)[index+1:]...)
				break
			}
		}
		if len(*streams) == 0 {
			log.Println("All clients gone. Stop casting.")
			delete(c.streams, source)
			stop = true
		}
	}()

	if !exists {
		log.Println("No active stream for source. Creating.")
		go func() {
			log.Println("Running ffmpeg for ", source)
			cmd := exec.Command("bash", "-c", "ffmpeg -i "+source+" -c:v mjpeg  -q:v 3 -huffman optimal -f mjpeg - 2>/dev/null")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Println("Error:", err)
				doneBroadcast()
				return
			}
			if err := cmd.Start(); err != nil {
				log.Println("Error:", err)
				doneBroadcast()
				return
			}
			buf := make([]byte, 512*1024)

			for !stop {
				n, err := stdout.Read(buf)
				if n == 0 || (err != nil && err != io.EOF) {
					doneBroadcast()
					log.Println("Error:", err)
					return
				}

				c.Lock()
				for _, cts := range *streams {
					cts.DataChan <- buf[:n]
				}
				c.Unlock()
			}
			doneBroadcast()
		}()
	} else {
		log.Println("Source exists. Attaching.")
	}

	return current.DataChan, current.DoneChan
}
