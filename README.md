# Docktor

Docktor is a little service that restarts unhealthy containers.

## Quck Start
```
docker pull vterdunov/docktor
docker run -d --rm --name=docktor -v /var/run/docker.sock:/var/run/docker.sock:ro vterdunov/docktor
```

## Configuration
Docktor read environment variables

| Variable | Values | Description |
|----------|--------|-------------|
| BACKOFF_JITTER | bool | Enable/Disable backoff jitter. Default is: `false` |
| BACKOFF_MIN_TIME | int | Sets the minimum delay time between restart container. Deafult is: `3`s |
| BACKOFF_MAX_TIME | int | Sets the maximum delay time between restart container. Default is: `30`s |
| DOCKER_HOST | string | Sets a path to docker daemon socket. Can be a unix ot tcp socket (`tcp://example.com:4243`, `unix:///var/run/docker.sock`).  Default is: `unix:///var/run/docker.sock` |

## Build
- Install Golang, Docker and [dep](https://github.com/golang/dep)  
- Ensure that dependencies was installed:  
`make dep`  
- Build docker container:  
`make build`  
- Or single binary file:  
`make compile`

## Run
`docker run -d --rm --name=docktor -v /var/run/docker.sock:/var/run/docker.sock:ro docktor`

### Example Docktor working process
`docker-compose up [--build] [--scale unhealthy=NUM ]`
