run: build
	@./bin/server

build:
	@go build -o bin/server .
