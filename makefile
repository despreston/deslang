all: build-server build-cli

build-server:
	go build -o bin/deslang-server cmd/deslang-server/server.go

build-cli:
	go build -o bin/deslang cmd/cli/cli.go

build-rpi-all: build-rpi-cli build-rpi-server

build-rpi-server:
	env GOOS=linux GOARCH=arm GOARM=7 go build -o bin/rpi/deslang-server cmd/deslang-server/server.go

build-rpi-cli:
	env GOOS=linux GOARCH=arm GOARM=7 go build -o bin/rpi/deslang cmd/cli/cli.go

