NAME=kubeval
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
OUTPUT ?= bin/darwin/amd64/$(NAME)

all: build

tools:
	git clone --depth 1 https://github.com/sstephenson/bats.git
	go get -u github.com/Masterminds/glide
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck

deps:
	glide install

check:
	errcheck

dirs:
	mkdir -p releases
	mkdir -p bin/linux/amd64
	mkdir -p bin/windows/amd64
	mkdir -p bin/darwin/amd64

build_deps: deps dirs

build: darwin linux windows

darwin: build_deps
	go build -v -o $(CURDIR)/${OUTPUT}
	tar -cvzf releases/$(NAME)-darwin-amd64.tar.gz bin/darwin/amd64/$(NAME)

linux: build_deps
	env GOOS=linux GOAARCH=amd64 go build -v -o $(CURDIR)/bin/linux/amd64/$(NAME)
	tar -cvzf releases/$(NAME)-linux-amd64.tar.gz bin/linux/amd64/$(NAME)

windows: build_deps
	env GOOS=windows GOAARCH=amd64 go build -v -o $(CURDIR)/bin/windows/amd64/$(NAME)
	tar -cvzf releases/$(NAME)-windows-amd64.tar.gz bin/windows/amd64/$(NAME)

example: darwin
	cat ./${OUTPUT} fixtures/valid.yaml

lint:
	golint

test:
	go test

acceptance:
	PATH=bin/darwin/amd64:$$PATH ./acceptance.bats

cover:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

clean:
	rm -fr releases bin

fmt:
	gofmt -w $(GOFMT_FILES)
