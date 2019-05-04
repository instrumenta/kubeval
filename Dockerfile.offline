FROM golang:1.12-alpine as builder
RUN apk --no-cache add make git
WORKDIR /
COPY . /
RUN make build

FROM alpine:latest as schemas
RUN apk --no-cache add git
RUN git clone --depth 1 https://github.com/instrumenta/kubernetes-json-schema.git
RUN git clone --depth 1 https://github.com/garethr/openshift-json-schema.git

FROM schemas as standalone-schemas
RUN cd kubernetes-json-schema/master && \
    find -maxdepth 1 -type d -not -name "." -not -name "*-standalone*" | xargs rm -rf

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/kubeval .
COPY --from=standalone-schemas /kubernetes-json-schema /schemas/kubernetes-json-schema/master
COPY --from=standalone-schemas /openshift-json-schema /schemas/openshift-json-schema/master
ENV KUBEVAL_SCHEMA_LOCATION=file:///schemas
RUN ln -s /kubeval /usr/local/bin/kubeval
ENTRYPOINT ["/kubeval"]
CMD ["--help"]
