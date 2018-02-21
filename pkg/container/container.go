package container

import (
	"fmt"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/jpillora/backoff"
	log "github.com/sirupsen/logrus"
)

type Patient struct {
	cid     string
	delay   time.Duration
	attempt int
	backoff backoff.Backoff
}

func init() {
	log.SetOutput(os.Stdout)
	if os.Getenv("DOCKTOR_LOGLEVEL") == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
}

// NewDockerClient creates a new Docker Clinet
func NewDockerClient() (*docker.Client, error) {
	if os.Getenv("DOCKER_HOST") == "" {
		client, err := docker.NewClient("unix:///var/run/docker.sock")
		if err != nil {
			return nil, fmt.Errorf("Error while get Docker client: %s", err)
		}
		return client, nil
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("Error while get Docker client: %s", err)
	}
	return client, nil
}

// Sorter sends unhealthy Container IDs into channel
func Sorter(events <-chan *docker.APIEvents, cid chan<- string) {
	for event := range events {
		switch event.Status {
		case "health_status: unhealthy":
			log.WithFields(log.Fields{
				"container_id": event.ID,
			}).Info("Found unhealthy container")

			cid <- event.ID
		}
	}
}

// Scheduler determines restart delay for container
func Scheduler(in <-chan string, out chan<- Patient, backoffMin, backoffMax time.Duration) {
	patients := make(map[string]Patient)
	var p Patient

	for containerID := range in {
		// check is it a new container or not
		if _, ok := patients[containerID]; ok {
			log.Debugf("I've already seen the container %s\n", containerID)
			p = patients[containerID]
			p.delay = p.backoff.Duration()
		} else {
			log.Debugf("I've never seen the container before %s\n", containerID)
			p = Patient{}
			p.cid = containerID

			// min =
			b := newBackoff(backoffMin, backoffMax)
			p.backoff = *b
			p.delay = b.Duration()
		}
		p.attempt++
		patients[containerID] = p
		log.WithFields(log.Fields{
			"function":     "scheduler",
			"attempt":      p.attempt,
			"delay":        p.delay,
			"container_id": containerID,
		}).Debug("Patient scheduled")

		// TODO: Implement delete containers from map

		out <- p
	}
}

// Restarter receives a patient and restart them
func Restarter(in <-chan Patient, client *docker.Client) {
	for p := range in {
		go func(p Patient) {
			cont, _ := client.InspectContainer(p.cid)
			if cont.State.Health.Status == "unhealthy" {
				log.Infof("Sleeping %s before restart", p.delay)
				time.Sleep(p.delay)

				log.WithFields(log.Fields{
					"function":       "restarter",
					"attempt":        p.attempt,
					"delay":          p.delay,
					"container_id":   p.cid,
					"container_name": cont.Name,
				}).Debug("Healing patient")

				err := client.RestartContainer(p.cid, 10)
				if err != nil {
					log.Errorf("Error restarting container: %s", p.cid)
				}
			}
		}(p)
	}
}

// PushAlredUnhealhyToScheduler pushes already running unhealthy containers
// into scheduler.
func PushAlredUnhealhyToScheduler(client *docker.Client, cid chan<- string) {
	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		cont, err := client.InspectContainer(container.ID)
		if err != nil {
			log.Warn("Cannot get container")
		}
		if cont.State.Health.Status == "unhealthy" {
			log.Infof("Container %s unhealthy. Healing...", container.ID)
			err = client.RestartContainer(container.ID, 10)
			if err != nil {
				log.Infof("Error restarting container: %s", container.ID)
			}
			log.Debugf("Restart container %s. OK", container.ID)
		}
	}
}

// newBackoff creates a new backoff instance
func newBackoff(min, max time.Duration) *backoff.Backoff {
	b := backoff.Backoff{
		Min:    min,
		Max:    max,
		Factor: 2,
		Jitter: true,
	}
	return &b
}
