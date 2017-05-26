rerun: build run

build:
	docker build -t docktor .

run:
	docker run -it --rm --name=docktor -v /var/run/docker.sock:/var/run/docker.sock:ro -m 500m --cpus=".5" docktor
