MODULE=rtsp2mjpeg

.PHONY: build run image

listen?=80


build: image
	docker run \
		-it \
		--rm \
		-v $(PWD)/src:/opt/src \
		-v $(PWD)/bin:/opt/bin \
		-e "GOPATH=/opt/go" \
		$(MODULE) \
		bash -c "cd /opt/src && go get -d -v && go build -i -o /opt/bin/$(MODULE) *.go"

run: stop image
	@docker run \
		-d \
		-t \
		--rm \
		-v $(PWD)/bin/$(MODULE):/opt/bin/$(MODULE) \
		-p "$(listen):80" \
		--name $(MODULE) \
		$(MODULE) \
		bash -c "cd /opt/bin && ./$(MODULE)"

stop:
	-@docker rm -f $(MODULE)

image:
	@docker build docker -t $(MODULE)
