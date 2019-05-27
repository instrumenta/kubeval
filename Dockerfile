FROM golang:1.12-alpine as builder
RUN apk --no-cache add make git
WORKDIR /
COPY . /
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/kubeval .
RUN ln -s /kubeval /usr/local/bin/kubeval
ENTRYPOINT ["/kubeval"]
CMD ["--help"]
