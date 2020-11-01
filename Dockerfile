FROM golang:alpine

COPY . /root/sparrow-scrobbler-server
WORKDIR /root/sparrow-scrobbler-server
RUN apk add libc-dev gcc git
RUN apk add -U --no-cache ca-certificates
RUN go get -d -v
RUN go build -ldflags "-linkmode external -extldflags -static"

FROM scratch
COPY --from=0 /root/sparrow-scrobbler-server/sparrow-scrobbler-server /sparrow-scrobbler-server
COPY --from=0 /root/sparrow-scrobbler-server/html /html
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/sparrow-scrobbler-server", "-v", "2", "-logtostderr"]
