PROG_NAME = docktor
GO_VARS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO_LDFLAGS=-v -ldflags="-s -w"

start: build run

build:
	docker build -t $(PROG_NAME) .

run:
	docker run -it --rm --name=$(PROG_NAME) -e DOCKTOR_LOGLEVEL=debug -v /var/run/docker.sock:/var/run/docker.sock:ro -m 500m --cpus=".5" $(PROG_NAME)

dep:
	@dep ensure -v

start-debug:
	docker build --tag $(PROG_NAME)-debug --file Dockerfile.debug .
	docker run -it --rm --name=$(PROG_NAME) -p 2345:2345 -v /var/run/docker.sock:/var/run/docker.sock:ro -m 500m --cpus=".5" $(PROG_NAME)-debug

compile:
	rm docktor 2>&1 || true
	$(GO_VARS) go build $(GO_LDFLAGS) -o $(PROG_NAME) ./cmd/docktor
