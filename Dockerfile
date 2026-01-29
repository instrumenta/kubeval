FROM golang:1.15-alpine as builder
RUN apk --no-cache add make git
WORKDIR /
COPY . /
RUN make build

FROM alpine:3.23.3
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/kubeval .
RUN ln -s /kubeval /usr/local/bin/kubeval
ENTRYPOINT ["/kubeval"]
CMD ["--help"]
