.PHONY: build clean deploy

build: 
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/main main.go

clean: 
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	STAGE=$(stage) sls deploy
