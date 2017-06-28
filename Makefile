NAME=kubeval
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
OUTPUT ?= bin/darwin/amd64/$(NAME)

all: build

$(GOPATH)/bin/glide:
	go get github.com/Masterminds/glide

$(GOPATH)/bin/golint:
	go get github.com/golang/lint/golint

$(GOPATH)/bin/errcheck:
	go get -u github.com/kisielk/errcheck

.bats:
	git clone --depth 1 https://github.com/sstephenson/bats.git .bats

vendor: glide.yaml $(GOPATH)/bin/glide
	glide install

check: $(GOPATH)/bin/errcheck
	errcheck

releases:
	mkdir -p releases

bin/linux/amd64:
	mkdir -p bin/linux/amd64

bin/windows/amd64:
	mkdir -p bin/windows/amd64

bin/darwin/amd64:
	mkdir -p bin/darwin/amd64

build: darwin linux windows

darwin: vendor releases bin/darwin/amd64
	go build -v -o $(CURDIR)/${OUTPUT}
	tar -cvzf releases/$(NAME)-darwin-amd64.tar.gz bin/darwin/amd64/$(NAME)

linux: vendor releases bin/linux/amd64
	env GOOS=linux GOAARCH=amd64 go build -v -o $(CURDIR)/bin/linux/amd64/$(NAME)
	tar -cvzf releases/$(NAME)-linux-amd64.tar.gz bin/linux/amd64/$(NAME)

windows: vendor releases bin/windows/amd64
	env GOOS=windows GOAARCH=amd64 go build -v -o $(CURDIR)/bin/windows/amd64/$(NAME)
	tar -cvzf releases/$(NAME)-windows-amd64.tar.gz bin/windows/amd64/$(NAME)

example: darwin
	./${OUTPUT} fixtures/valid.yaml

lint: $(GOPATH)/bin/golint
	golint

test:
	go test

acceptance: .bats
	env PATH=./.bats/bin:$$PATH:./bin/darwin/amd64 ./acceptance.bats

cover:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

clean:
	rm -fr releases bin

fmt:
	gofmt -w $(GOFMT_FILES)

.PHONY: fmt clean cover acceptance test example windows linux darwin build check
