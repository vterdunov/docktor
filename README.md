# Docktor

Listen Docker events and restarting unhealthy containers.

## Build
`docker build -t docktor .`

## Run
`docker run -d --rm --name=docktor -v /var/run/docker.sock:/var/run/docker/sock:ro docktor`
