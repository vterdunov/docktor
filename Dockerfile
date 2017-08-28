FROM golang:1.8.3-alpine AS build-stage

ARG WORKDIR=/go/src/github.com/vterdunov/docktor

WORKDIR $WORKDIR

RUN apk add --no-cache git build-base
RUN go get -v github.com/golang/dep/cmd/dep

COPY . $WORKDIR
RUN [ -d 'vendor' ] || make dep

RUN make compile

FROM scratch
CMD ["/docktor"]
COPY --from=build-stage /go/src/github.com/vterdunov/docktor/docktor /docktor
