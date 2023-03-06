NAME=kubeval
IMAGE_NAME=garethr/$(NAME)
PACKAGE_NAME=github.com/instrumenta/$(NAME)
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
TAG=$(shell git describe --abbrev=0 --tags)


all: help

$(GOPATH)/bin/golint$(suffix):
	go get github.com/golang/lint/golint

$(GOPATH)/bin/goveralls$(suffix):
	go get github.com/mattn/goveralls


## build: Build the project
build: bin
	@go build -o bin/$(NAME) .

## vandor:  verify dependencies have expected conten
vendor:
	@go mod vendor


.bats:
	@git clone --depth 1 https://github.com/sstephenson/bats.git .bats

## bin: Create the bin directory
bin:
	@mkdir bin

## release: Create a release
release:
	@goreleaser --rm-dist 

## snapshot: Creating a Snapshot Version
snapshot:
	@goreleaser --snapshot --skip-publish --rm-dist

## lint: Use golint to check the code specification
lint: $(GOPATH)/bin/golint$(suffix)
	@golint

## docker: Build and push the docker image
docker:
	@docker build -t $(IMAGE_NAME):$(TAG) .
	@docker tag $(IMAGE_NAME):$(TAG) $(IMAGE_NAME):latest
	@docker push $(IMAGE_NAME):$(TAG)
	@docker push $(IMAGE_NAME):latest

## docker-offline: Build and push an offline docker image
docker-offline:
	@docker build -f Dockerfile.offline -t $(IMAGE_NAME):$(TAG)-offline .
	@docker tag $(IMAGE_NAME):$(TAG)-offline $(IMAGE_NAME):offline

## vet: Check the code using go vet
vet:
	@go vet

## test: Run the tests
test: vet
	@go test -race -v -cover ./...

## watch: Monitor code changes and run tests automatically
watch:
	@ls */*.go | entr make test

## acceptance: Operational acceptance test
acceptance:
	@docker build -f Dockerfile.acceptance -t $(IMAGE_NAME):$(TAG)-acceptance .
	@docker tag $(IMAGE_NAME):$(TAG)-acceptance $(IMAGE_NAME):acceptance

## cover: Generate code coverage reports
cover:
	@go test -v ./$(NAME) -coverprofile=coverage.out
	@go tool cover -html=coverage.out
	@rm coverage.out

## clean: Clear the generated file
clean:
	@rm -fr dist bin

## fmt: Format the code
fmt:
	@gofmt -w $(GOFMT_FILES)

## check: Generates the checksum of the file
dist/$(NAME)-checksum-%:
	@cd dist && sha256sum $@.zip

checksums: dist/$(NAME)-checksum-darwin-amd64 dist/$(NAME)-checksum-windows-386 dist/$(NAME)-checksum-windows-amd64 dist/$(NAME)-checksum-linux-amd64

chocolatey/$(NAME)/$(NAME).$(TAG).nupkg: chocolatey/$(NAME)/$(NAME).nuspec
	@cd chocolatey/$(NAME) && choco pack

## choco: Build and push the chocolatey package
choco:
	@cd chocolatey/$(NAME) && choco push $(NAME).$(TAG).nupkg -s https://chocolatey.org/

## help: Display help information
help: Makefile
	@echo ""
	@echo "Usage:"
	@echo ""
	@echo "  make [target]"
	@echo ""
	@echo ""
	@echo "Targets:"
	@echo ""
	@awk -F ':|##' '/^[^\.%\t][^\t]*:.*##/{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST) | sort
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: release snapshot fmt clean cover acceptance lint docker test vet watch build check choco checksums
