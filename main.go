package main

import (
	"log"
	"time"

	"github.com/fsouza/go-dockerclient"
)

const restartTimeout uint = 10

type patientCard struct {
	cid   string
	delay time.Duration
}

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
	log.Println("Listening for Docker events...")

	// Restart already running containers.
	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		cont, _ := client.InspectContainer(container.ID)
		if cont.State.Health.Status == "unhealthy" {
			log.Printf("Container %s unhealthy. Healing...", container.ID)
			err = client.RestartContainer(container.ID, restartTimeout)
			if err != nil {
				log.Printf("Error restarting container: %s", container.ID)
			}
		}
	}

	// Listen events
	unhealthy := make(chan string)
	patients := make(chan patientCard)

	go sorter(events, unhealthy)
	go scheduler(unhealthy, patients)
	restarter(patients, client)
}

// Sends unhealthy Container ID into channel
func sorter(in <-chan *docker.APIEvents, out chan<- string) {
	for event := range in {
		switch event.Status {
		case "health_status: unhealthy":
			log.Printf("Found unhealthy container: %s", event.ID)
			out <- event.ID
		}
	}
}

// Determines restart delay for container
func scheduler(in <-chan string, out chan<- patientCard) {
	counter := make(map[string]uint)
	for containerID := range in {
		log.Printf("Scheduling container: %s", containerID)
		_, exist := counter[containerID]
		if exist {
			if counter[containerID] < 4 {
				counter[containerID]++
			}
		} else {
			counter[containerID] = 1
		}

		// TODO: Implement exponential backoff algoritm
		// TODO: Implement delete containers from map
		var delay time.Duration
		switch counter[containerID] {
		case 1:
			delay = 10
		case 2:
			delay = 15
		case 3:
			delay = 30
		case 4:
			delay = 60
		default:
			delay = 60
		}
		p := patientCard{
			cid:   containerID,
			delay: delay,
		}
		out <- p
	}
}

// Receives a patientCard and restart him
func restarter(in <-chan patientCard, client *docker.Client) {
	for p := range in {
		go func(p patientCard) {
			log.Printf("Sleeping %d sec before restart %s", p.delay, p.cid)
			time.Sleep(p.delay * time.Second)
			cont, _ := client.InspectContainer(p.cid)
			if cont.State.Health.Status == "unhealthy" {
				log.Printf("Healing patient: %s", p.cid)
				err := client.RestartContainer(p.cid, restartTimeout)
				if err != nil {
					log.Printf("Error restarting container: %s", p.cid)
				}
			}
		}(p)
	}
}
