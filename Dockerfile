FROM golang:1.15-alpine as builder
RUN apk --no-cache add make git
WORKDIR /
COPY . /
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/kubeval .
RUN ln -s /kubeval /usr/local/bin/kubeval
RUN printf '#!/bin/sh\n\neval kubeval "$@"' > /kubeval.sh
RUN chmod +x /kubeval.sh
ENTRYPOINT ["/kubeval.sh"]
CMD ["--help"]
