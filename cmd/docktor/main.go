package main

import (
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/go-kit/kit/log"
	"github.com/vterdunov/docktor/pkg/config"
	"github.com/vterdunov/docktor/pkg/container"
)

func main() {
	cfg := config.NewConfig()
	var l log.Logger
	l = log.NewJSONLogger(os.Stderr)
	l = log.With(l, "ts", log.DefaultTimestampUTC)
	l = log.With(l, "caller", log.DefaultCaller)

	client, err := container.NewDockerClient()
	if err != nil {
		l.Log("err", "could not create Docker client")
		os.Exit(1)
	}

	dockerEvents := make(chan *docker.APIEvents)
	err = client.AddEventListener(dockerEvents)
	if err != nil {
		l.Log("err", err)
		os.Exit(1)
	}
	l.Log("msg", "Listening for Docker events...")

	unhealthyCIDs := make(chan string, 10)
	patients := make(chan container.Patient, 10)

	go container.Sorter(dockerEvents, unhealthyCIDs)
	go container.Scheduler(unhealthyCIDs, patients, cfg.BackoffMinTime, cfg.BackoffMaxTime, cfg.BackoffJitter, log.With(l, "component", "scheduler"))

	container.PushAlredUnhealhy(client, unhealthyCIDs, log.With(l, "component", "pusher"))

	container.Restarter(patients, client, log.With(l, "component", "restarter"))
}
