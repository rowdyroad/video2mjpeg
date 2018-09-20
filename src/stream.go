package main

import "image"

type Stream struct {
	ImageChan chan image.Image
	DoneChan  chan bool
}
