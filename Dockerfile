FROM golang:1.8-alpine as builder
RUN apk --no-cache add make git
RUN mkdir -p /go/src/github.com/garethr/kubeval/
COPY . /go/src/github.com/garethr/kubeval/
WORKDIR /go/src/github.com/garethr/kubeval/
RUN make linux

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/garethr/kubeval/bin/linux/amd64/kubeval .
ENTRYPOINT ["/kubeval"]
CMD ["--help"]
