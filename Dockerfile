FROM golang:1.8.3-alpine AS build-stage

WORKDIR /build
RUN apk update
RUN apk add --no-cache git build-base
RUN go get github.com/fsouza/go-dockerclient
COPY . /build
RUN go build -ldflags="-s -w" -o docktor main.go

FROM alpine:3.5
CMD ["/docktor"]
COPY --from=build-stage /build/docktor .
