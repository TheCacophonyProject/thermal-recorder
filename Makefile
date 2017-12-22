VERSION := $(shell git describe --always --dirty)

release:
	go build -ldflags="-X main.version=${VERSION}" 
