package caster

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"strings"
	"sync"

	log "github.com/rowdyroad/go-simple-logger"
)

//NewCaster creates Caster instance
func NewCaster() *Caster {
	return &Caster{
		sources: map[string]*Source{},
	}
}

//Caster converting rtsp(or any) stream to mjpeg (with ffmjpeg support)
type Caster struct {
	sync.Mutex
	sources map[string]*Source
}

//Close all sourcess
func (c *Caster) Close() {
	c.Lock()
	defer c.Unlock()
	for _, source := range c.sources {
		source.Close()
	}
}

//Cast main functionality for convert rtsp(or any) to mjpeg
func (c *Caster) Cast(input string, fps *int64, qscale *int64, scale *string, stopChan <-chan bool) (chan image.Image, chan bool, error) {
	c.Lock()
	defer c.Unlock()

	id := input
	if fps != nil {
		id +=fmt.Sprintf("fps=%d", *fps)
	}
	if qscale != nil {
		id +=fmt.Sprintf("qscale=%d", *qscale)
	}
	if scale != nil {
		id +=fmt.Sprintf("scale=%s", *scale)
	}

	log.Debug("Casting for", id)
	var exists bool
	var source *Source
	source, exists = c.sources[id]
	if !exists {
		source = &Source{sync.Mutex{}, []*Stream{}, nil, false}
		c.sources[id] = source
	}

	stream := &Stream{make(chan image.Image), make(chan bool)}
	source.Lock()
	source.Streams = append(source.Streams, stream)
	source.Unlock()

	go c.waitForClientGone(source, stream, stopChan)

	if !exists {
		go c.broadcastSource(id, input , fps, qscale, scale , source)
	} else {
		log.Debug("Source exists. Attaching.")
	}

	return stream.ImageChan, stream.DoneChan, nil
}

func (c *Caster) waitForClientGone(source *Source, stream *Stream, stopChan <-chan bool) {
	<-stopChan
	log.Debug("Client gone")

	<-stream.ImageChan

	source.Lock()
	log.Debug("Find and remove stream from source");
	for index, cts := range source.Streams {
		if cts == stream {
			log.Debug("Removing client stream record.")
			source.Streams = append(source.Streams[:index], source.Streams[index+1:]...)
			break
		}
	}
	if len(source.Streams) == 0 {
		log.Debug("All clients gone. Source closing.")
		source.Close()
	}
	source.Unlock()

	log.Debug("Source is stopped:", source.Stop)
	if source.Stop {
		log.Debug("Remove stopped source from sources list")
		c.Lock()
		for id, s := range c.sources {
			if s == source {
				log.Debug("Source closed. Deleting source", id)
				delete(c.sources, id)
				break
			}
		}
		c.Unlock()
		log.Debug("done allclient.")
	}
}

func (c *Caster) broadcastSource(id string, input string, fps *int64, qscale *int64, scale *string, source *Source) {
	log.Debug("No active stream for source. Creating.")
	defer source.Close()

	log.Debug("Running ffmpeg for", id)

	input = strings.ReplaceAll(input,"\"", "\\\"", )

	execCommand := "ffmpeg -i \"" + input + "\" -c:v mjpeg -f mjpeg"
	if fps !=nil {
		execCommand += fmt.Sprintf(" -r %d ", *fps)
	}

	if qscale != nil {
		execCommand += fmt.Sprintf(" -q:v %d ", *qscale)
	}

	if scale != nil {
		execCommand += fmt.Sprintf(" -vf 'scale=%s' ", strings.ReplaceAll(strings.ReplaceAll(*scale, "'", "\\'", ), "\"", "\\\"", ))
	}

	log.Debug("Exec command:", execCommand)
	cmd := exec.Command("bash", "-c", execCommand+" -")
	var err error
	source.Pipe, err = cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	if err != nil {
		log.Error("Error:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Error("Error:", err)
		return
	}

	for !source.Stop {
		if image, err := jpeg.Decode(source.Pipe); err == nil {
			source.Lock()
			for _, stream := range source.Streams {
				stream.ImageChan <- image
			}
			source.Unlock()
		}
	}

	log.Debug("Stopped")
}
