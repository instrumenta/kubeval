NAME=kubeval
IMAGE_NAME=garethr/$(NAME)
PACKAGE_NAME=github.com/garethr/$(NAME)
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
TAG=$(shell git describe --abbrev=0 --tags)

LDFLAGS += -X "$(PACKAGE_NAME)/version.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "$(PACKAGE_NAME)/version.BuildVersion=$(shell git describe --abbrev=0 --tags)"
LDFLAGS += -X "$(PACKAGE_NAME)/version.BuildSHA=$(shell git rev-parse HEAD)"
# Strip debug information
LDFLAGS += -s

ifeq ($(OS),Windows_NT)
	suffix := .exe
endif

all: build

$(GOPATH)/bin/glide$(suffix):
	go get github.com/Masterminds/glide

$(GOPATH)/bin/golint$(suffix):
	go get github.com/golang/lint/golint

$(GOPATH)/bin/goveralls$(suffix):
	go get github.com/mattn/goveralls

$(GOPATH)/bin/errcheck$(suffix):
	go get -u github.com/kisielk/errcheck

glide: $(GOPATH)/bin/glide$(suffix)

.bats:
	git clone --depth 1 https://github.com/sstephenson/bats.git .bats

glide.lock: glide.yaml glide
	glide update

vendor: glide.lock
	glide install

check: vendor $(GOPATH)/bin/errcheck$(suffix)
	errcheck

releases:
	mkdir -p releases

bin/linux/amd64:
	mkdir -p bin/linux/amd64

bin/windows/amd64:
	mkdir -p bin/windows/amd64

bin/windows/386:
	mkdir -p bin/windows/386

bin/darwin/amd64:
	mkdir -p bin/darwin/amd64

build: darwin linux windows

darwin: vendor releases bin/darwin/amd64
	env CGO_ENABLED=0 GOOS=darwin GOAARCH=amd64 go build -ldflags '$(LDFLAGS)' -v -o $(CURDIR)/bin/darwin/amd64/$(NAME)
	tar -C bin/darwin/amd64 -cvzf releases/$(NAME)-darwin-amd64.tar.gz $(NAME)

linux: vendor releases bin/linux/amd64
	env CGO_ENABLED=0 GOOS=linux GOAARCH=amd64 go build -ldflags '$(LDFLAGS)' -v -o $(CURDIR)/bin/linux/amd64/$(NAME)
	tar -C bin/linux/amd64 -cvzf releases/$(NAME)-linux-amd64.tar.gz $(NAME)

windows: windows-64 windows-32

windows-64: vendor releases bin/windows/amd64
	env CGO_ENABLED=0 GOOS=windows GOAARCH=amd64 go build -ldflags '$(LDFLAGS)' -v -o $(CURDIR)/bin/windows/amd64/$(NAME).exe
	tar -C bin/windows/amd64 -cvzf releases/$(NAME)-windows-amd64.tar.gz $(NAME).exe
	cd bin/windows/amd64 && zip ../../../releases/$(NAME)-windows-amd64.zip $(NAME).exe

windows-32: vendor releases bin/windows/386
	env CGO_ENABLED=0 GOOS=windows GOAARCH=386 go build -ldflags '$(LDFLAGS)' -v -o $(CURDIR)/bin/windows/386/$(NAME).exe
	tar -C bin/windows/386 -cvzf releases/$(NAME)-windows-386.tar.gz $(NAME).exe
	cd bin/windows/386 && zip ../../../releases/$(NAME)-windows-386.zip $(NAME).exe

lint: $(GOPATH)/bin/golint$(suffix)
	golint

docker:
	docker build -t $(IMAGE_NAME):$(TAG) .
	docker tag $(IMAGE_NAME):$(TAG) $(IMAGE_NAME):latest

docker-offline:
	docker build -f Dockerfile.offline -t $(IMAGE_NAME):$(TAG)-offline .
	docker tag $(IMAGE_NAME):$(TAG)-offline $(IMAGE_NAME):offline

publish: docker docker-offline
	docker push $(IMAGE_NAME):$(TAG)
	docker push $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):$(TAG)-offline
	docker push $(IMAGE_NAME):offline

vet: glide
	go vet $(shell glide novendor)

test: vendor vet lint check glide
	go test -race -v -cover $(shell glide novendor)

coveralls: vendor $(GOPATH)/bin/goveralls$(suffix)
	goveralls -service=travis-ci

watch:
	ls */*.go | entr make test

acceptance:
	docker build -f Dockerfile.acceptance -t $(IMAGE_NAME):$(TAG)-acceptance .
	docker tag $(IMAGE_NAME):$(TAG)-acceptance $(IMAGE_NAME):acceptance
	docker run --rm -t $(IMAGE_NAME):acceptance

cover:
	go test -v ./$(NAME) -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

clean:
	rm -fr releases bin

fmt:
	gofmt -w $(GOFMT_FILES)

checksum-windows-386:
	cd releases && powershell -Command "(Get-FileHash $(NAME)-windows-386.zip -Algorithm SHA256).Hash.ToLower()"

checksum-windows-amd64:
	cd releases && powershell -Command "(Get-FileHash $(NAME)-windows-amd64.zip -Algorithm SHA256).Hash.ToLower()"

checksum-darwin:
	cd releases && powershell -Command "(Get-FileHash $(NAME)-darwin-amd64.tar.gz -Algorithm SHA256).Hash.ToLower()"

checksum-linux:
	cd releases && powershell -Command "(Get-FileHash $(NAME)-linux-amd64.tar.gz -Algorithm SHA256).Hash.ToLower()"

checksums: download checksum-darwin checksum-windows-386 checksum-windows-amd64 checksum-linux

chocolatey/$(NAME)/$(NAME).$(TAG).nupkg: chocolatey/$(NAME)/$(NAME).nuspec
	cd chocolatey/$(NAME) && choco pack

choco:
	cd chocolatey/$(NAME) && choco push $(NAME).$(TAG).nupkg -s https://chocolatey.org/

download:
	powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iwr -outf releases/$(NAME)-darwin-amd64.tar.gz https://$(PACKAGE_NAME)/releases/download/$(TAG)/$(NAME)-darwin-amd64.tar.gz"
	powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iwr -outf releases/$(NAME)-windows-amd64.zip https://$(PACKAGE_NAME)/releases/download/$(TAG)/$(NAME)-windows-amd64.zip"
	powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iwr -outf releases/$(NAME)-windows-386.zip https://$(PACKAGE_NAME)/releases/download/$(TAG)/$(NAME)-windows-386.zip"
	powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iwr -outf releases/$(NAME)-linux-amd64.tar.gz https://$(PACKAGE_NAME)/releases/download/$(TAG)/$(NAME)-linux-amd64.tar.gz"

.PHONY: fmt clean cover acceptance lint docker test vet watch windows linux darwin build check checksum-windows-386 checksum-windows-amd64 checksum-darwin checksum-linux choco download checksum
