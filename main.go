package main

import (
	"log"

	"github.com/fsouza/go-dockerclient"
)

var restartTimeout uint = 10

func main() {
	endpoint := "unix:///var/run/docker.sock"
	// TODO: use FromEnv endpoint
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	events := make(chan *docker.APIEvents)
	err = client.AddEventListener(events)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening for Docker events ...")

	defer func() {
		err = client.RemoveEventListener(events)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// TODO: Heal already running containers too
	for msg := range events {
		switch msg.Status {
		case "health_status: unhealthy":
			log.Printf("Container %s unhealthy. Healing...", msg.ID)
			err = client.RestartContainer(msg.ID, restartTimeout)
			if err != nil {
				log.Fatalf("Error restarting container: %s", msg.ID)
			}
		}
	}
}
