install:
	cd ./xfsvolctl && go install -v

test:
	cd ./manager && go test -v

fmt:
	cd ./manager && go fmt
	cd ./lib && go fmt
	cd ./main && go fmt
	cd ./xfsvolctl && go fmt

.PHONY: install
