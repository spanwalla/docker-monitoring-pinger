include .env
export

compose-up: ### Run docker-compose
	docker-compose up --build -d && docker-compose logs -f
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down

install: ### Install to $GOPATH/bin
	go install github.com/spanwalla/docker-monitoring-pinger@latest
.PHONY: install

build: ### Build binaries for different platforms
	mkdir bin
	GOOS=linux GOARCH=amd64 go build -o bin/docker-pinger-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o bin/docker-pinger-macos-amd64 .
	GOOS=windows GOARCH=amd64 go build -o bin/docker-pinger-windows-amd64.exe .
.PHONY: build