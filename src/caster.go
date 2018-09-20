package main

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

//NewCaster creates Caster instance
func NewCaster() *Caster {
	return &Caster{sync.Map{}}
}

//Caster converting rtsp(or any) stream to mjpeg (with ffmjpeg support)
type Caster struct {
	sources sync.Map
}

func (c *Caster) Close() {
	c.sources.Range(func(k, source interface{}) bool {
		source.(*Source).Close(true)
		return true
	})
}

//Cast main functionality for convert rtsp(or any) to mjpeg
func (c *Caster) Cast(command map[string]string, stopChan <-chan bool) (chan image.Image, chan bool, error) {

	id := ""
	for name, value := range command {
		id += name + "=" + value + ";"
	}

	source, has := command["source"]
	if !has {
		return nil, nil, errors.New("source attribute is required")
	}
	fps, _ := strconv.ParseInt(command["fps"], 10, 64)
	qscale, _ := strconv.ParseInt(command["qscale"], 10, 64)
	scale, _ := command["scale"]

	log.Println("Casting for", id)

	s, exists := c.sources.Load(id)
	if !exists {
		s = &Source{sync.Mutex{}, []Stream{}, nil, false}
		c.sources.Store(id, s)
	}
	currentSource := s.(*Source)
	current := Stream{make(chan image.Image), make(chan bool)}
	currentSource.Streams = append(currentSource.Streams, current)

	go func() {
		<-stopChan
		log.Println("Client gone")
		currentSource.Lock()
		defer currentSource.Unlock()
		for index, cts := range currentSource.Streams {
			if cts == current {
				log.Println("Removing client stream record.")
				currentSource.Streams = append(currentSource.Streams[:index], currentSource.Streams[index+1:]...)
				break
			}
		}

		log.Println("Streams:", len(currentSource.Streams))

		if len(currentSource.Streams) == 0 {
			log.Println("All clients gone. Stop casting.")
			currentSource.Close(false)
			c.sources.Delete(id)
			log.Println("done allclient.")
		}
	}()

	if !exists {
		log.Println("No active stream for source. Creating.")
		go func() {
			defer currentSource.Close(true)

			log.Println("Running ffmpeg for", id)

			execCommand := "ffmpeg -i " + source + " -c:v mjpeg -f mjpeg"
			if fps > 0 {
				execCommand += fmt.Sprintf(" -r %d ", fps)
			}

			if qscale > 0 {
				execCommand += fmt.Sprintf(" -q:v %d ", qscale)
			}

			if len(scale) > 0 {
				execCommand += fmt.Sprintf(" -vf 'scale=%s' ", strings.Replace(scale, "'", "\\'", -1))
			}

			log.Println("Exec command:", execCommand)
			cmd := exec.Command("bash", "-c", execCommand+" - 2>/dev/null")
			var err error
			currentSource.Pipe, err = cmd.StdoutPipe()
			if err != nil {
				log.Println("Error:", err)
				return
			}

			if err := cmd.Start(); err != nil {
				log.Println("Error:", err)
				return
			}

			for !currentSource.Stop {
				if image, err := jpeg.Decode(currentSource.Pipe); err == nil {
					currentSource.Lock()
					for _, stream := range currentSource.Streams {
						stream.ImageChan <- image
					}
					currentSource.Unlock()
				} else {
					log.Println("Error image:", err)
				}
			}

			log.Println("Stopped")
		}()
	} else {
		log.Println("Source exists. Attaching.")
	}

	return current.ImageChan, current.DoneChan, nil
}
