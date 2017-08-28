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

### Run Test Unhealthy Container
`docker run -d --rm --health-cmd "exit 1" --health-interval=2s --health-timeout=3s --name unhealty1 alpine sleep 9999`
