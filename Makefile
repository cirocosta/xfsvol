all: build

build:
	cd ./xfsvolctl && go build -v

install:
	cd ./xfsvolctl && go install -v
	cd ./main && go install -v

test:
	cd ./manager && go test -v
	cd ./lib && go test -v

fmt:
	cd ./manager && go fmt
	cd ./lib && go fmt
	cd ./main && go fmt
	cd ./xfsvolctl && go fmt

.PHONY: all install test build fmt
