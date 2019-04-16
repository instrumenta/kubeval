FROM golang:1.12-alpine as builder
RUN apk --no-cache add make git
WORKDIR /
COPY . /
RUN make linux

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/linux/amd64/kubeval .
RUN ln -s /kubeval /usr/local/bin/kubeval
ENTRYPOINT ["/kubeval"]
CMD ["--help"]
