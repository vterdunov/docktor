# Docktor

Restarts unhealthy containers.

## Quck Start
```
docker pull vterdunov/docktor
docker run -d --rm --name=docktor -v /var/run/docker.sock:/var/run/docker.sock:ro vterdunov/docktor
```

## Build
`docker build -t docktor .`

## Run
`docker run -d --rm --name=docktor -v /var/run/docker.sock:/var/run/docker.sock:ro docktor`

### Example Docktor working process
`docker-compose up [--build] [--scale unhealthy=NUM ]`
