.PHONY: test
test:
	go test ./...

.PHONY: build-server-osx
build-server-osx:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/wall-fractal-osx cmd/server

.PHONY: build-server-rpi
build-server-rpi:
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -o dist/wall-fractal-rpi cmd/server
