# rtsp2mjpeg 
Module to casting mjpeg stream over http from rtsp (or any) video source


Instructions:
$ sudo apt install fmpeg

$ git clone https://github.com/rowdyroad/rtsp2mjpeg.git
$ cd rstp2mjpeg
$ go run *.go

Open you browser at http://localhost:8888/video?source=rtsp://184.72.239.149/vod/mp4:BigBuckBunny_175k.mov
