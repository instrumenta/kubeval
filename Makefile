NAME=kubeval
IMAGE_NAME=garethr/$(NAME)
PACKAGE_NAME=github.com/instrumenta/$(NAME)
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
TAG=$(shell git describe --abbrev=0 --tags)


all: build

$(GOPATH)/bin/golint$(suffix):
	go get github.com/golang/lint/golint

$(GOPATH)/bin/goveralls$(suffix):
	go get github.com/mattn/goveralls

vendor:
	go mod vendor

.bats:
	git clone --depth 1 https://github.com/sstephenson/bats.git .bats

bin:
	mkdir bin

release:
	goreleaser --rm-dist

snapshot:
	goreleaser --snapshot --skip-publish --rm-dist

build: bin
	go build -o bin/$(NAME) .

lint: $(GOPATH)/bin/golint$(suffix)
	golint

docker:
	docker build -t $(IMAGE_NAME):$(TAG) .
	docker tag $(IMAGE_NAME):$(TAG) $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):$(TAG)
	docker push $(IMAGE_NAME):latest

docker-offline:
	docker build -f Dockerfile.offline -t $(IMAGE_NAME):$(TAG)-offline .
	docker tag $(IMAGE_NAME):$(TAG)-offline $(IMAGE_NAME):offline

vet:
	go vet

test: vet
	go test -race -v -cover ./...

watch:
	ls */*.go | entr make test

acceptance:
	docker build -f Dockerfile.acceptance -t $(IMAGE_NAME):$(TAG)-acceptance .
	docker tag $(IMAGE_NAME):$(TAG)-acceptance $(IMAGE_NAME):acceptance

cover:
	go test -v ./$(NAME) -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

clean:
	rm -fr dist bin

fmt:
	gofmt -w $(GOFMT_FILES)

dist/$(NAME)-checksum-%:
	cd dist && sha256sum $@.zip

checksums: dist/$(NAME)-checksum-darwin-amd64 dist/$(NAME)-checksum-windows-386 dist/$(NAME)-checksum-windows-amd64 dist/$(NAME)-checksum-linux-amd64

chocolatey/$(NAME)/$(NAME).$(TAG).nupkg: chocolatey/$(NAME)/$(NAME).nuspec
	cd chocolatey/$(NAME) && choco pack

choco:
	cd chocolatey/$(NAME) && choco push $(NAME).$(TAG).nupkg -s https://chocolatey.org/

.PHONY: release snapshot fmt clean cover acceptance lint docker test vet watch build check choco checksums
