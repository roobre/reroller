FROM golang:alpine as builder

COPY . src
RUN cd src && CGO_ENABLED=0 go build -o /reroller ./cmd

FROM alpine:latest

RUN apk add tini
COPY --from=builder /reroller /usr/bin/

ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/reroller"]
