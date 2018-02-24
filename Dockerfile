FROM golang:1.10-alpine AS build-stage

WORKDIR /go/src/github.com/vterdunov/docktor

RUN apk add --no-cache git build-base
RUN go get -v github.com/golang/dep/cmd/dep

COPY . /go/src/github.com/vterdunov/docktor
RUN [ -d 'vendor' ] || make dep

RUN make compile

FROM scratch
CMD ["/docktor"]
COPY --from=build-stage /go/src/github.com/vterdunov/docktor/docktor /docktor
