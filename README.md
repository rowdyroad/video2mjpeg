# video2mjpeg
Module to casting mjpeg stream over http from any video source


## Direct using
Instructions:
```sh
$ sudo apt install fmpeg

$ git clone https://github.com/rowdyroad/video2mjpeg.git
$ cd rstp2mjpeg
$ go run src/*.go
```
Open http://localhost/video?source=rtsp://184.72.239.149/vod/mp4:BigBuckBunny_175k.mov in your browser.

## Dockerizable using
```sh
$ make build
$ make run
```

Open http://localhost/video?source=rtsp://184.72.239.149/vod/mp4:BigBuckBunny_175k.mov in your browser.

Start on different port:
```sh
$ make run listen=8000
```

## URL attributes
 - source - url for vidoe stream (like in an example)
 - fps - frames per second
 - qscale - quality of image (1-3) (read ffmpeg man)
 - scale - scale of image (640:480, 500:-1 & etc.) (read ffmpeg man "-vf scale")

Example: http://localhost/video?source=rtsp://184.72.239.149/vod/mp4:BigBuckBunny_175k.mov&fps=1&qscale=3&scale=640:480
