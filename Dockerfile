FROM golang:1.8.3-alpine AS build-stage

WORKDIR /build
RUN apk update && \
    apk add --no-cache git build-base
RUN go get -v github.com/fsouza/go-dockerclient
COPY main.go /build
RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-s -w" -o docktor main.go

FROM scratch
CMD ["/docktor"]
COPY --from=build-stage /build/docktor /docktor
