.PHONY: test
test:
	go test ./...

.PHONY: build-server-osx
build-server-osx:
	GO111MODULE=on GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/wall-fractal-osx ./cmd/server

.PHONY: build-server-rpi
build-server-rpi:
	GO111MODULE=on GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -o dist/wall-fractal-rpi ./cmd/server

.PHONY: mod-init
mod-init:
	GO111MODULE=on go mod init

.PHONY: mod-vendor
mod-vendor:
	GO111MODULE=on go mod vendor
