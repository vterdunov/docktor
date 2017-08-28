package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/jpillora/backoff"
)

const restartTimeout uint = 10

var (
	backoffMin, backoffMax time.Duration
)

type patient struct {
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

func main() {
	minBackoff := os.Getenv("DOCKTOR_BACKOFF_MIN_TIME")
	minTime, err := stringToTime(minBackoff, 5)
	if err != nil {
		log.Fatal(err)
	}
	backoffMin = minTime
	log.Infof("Minimal restart time: %s", backoffMin)

	maxBackoff := os.Getenv("DOCKTOR_BACKOFF_MAX_TIME")
	maxTime, err := stringToTime(maxBackoff, 300)
	if err != nil {
		log.Fatal(err)
	}
	backoffMax = maxTime
	log.Infof("Maximal restart time: %s", backoffMax)

	client, err := newDockerClient()
	if err != nil {
		log.Fatal(err)
	}

	dockerEvents := make(chan *docker.APIEvents)
	err = client.AddEventListener(dockerEvents)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Listening for Docker events...")

	// Listen events
	unhealthy := make(chan string, 10)
	patients := make(chan patient, 10)

	go sorter(dockerEvents, unhealthy)
	go scheduler(unhealthy, patients)

	// Restart already running containers.
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
			err = client.RestartContainer(container.ID, restartTimeout)
			if err != nil {
				log.Infof("Error restarting container: %s", container.ID)
			}
			log.Debugf("Restart container %s. OK", container.ID)
		}
	}

	restarter(patients, client)
}

// stringToTime converts a string to time.Duration in seconds
func stringToTime(envVar string, defaultTime uint) (time.Duration, error) {
	if envVar == "" {
		return time.Duration(defaultTime) * time.Second, nil
	}
	integer, err := strconv.Atoi(envVar)
	if err != nil {
		return 0, fmt.Errorf("Cannot parse: %s", envVar)
	}
	return time.Duration(integer) * time.Second, nil
}

// newDockerClient creates a new Docker Clinet
func newDockerClient() (*docker.Client, error) {
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

// sorter sends unhealthy Container ID into channel
func sorter(in <-chan *docker.APIEvents, out chan<- string) {
	for event := range in {
		switch event.Status {
		case "health_status: unhealthy":
			log.WithFields(log.Fields{
				"container_id": event.ID,
			}).Info("Found unhealthy container")

			out <- event.ID
		}
	}
}

// scheduler determines restart delay for container
func scheduler(in <-chan string, out chan<- patient) {
	patients := make(map[string]patient)
	var p patient

	for containerID := range in {
		// check is it a new container or not
		if _, ok := patients[containerID]; ok {
			log.Debugf("I've already seen the container %s\n", containerID)
			p = patients[containerID]
			p.delay = p.backoff.Duration()
		} else {
			log.Debugf("I've never seen the container before %s\n", containerID)
			p = patient{}
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

// restarter receives a patient and restart them
func restarter(in <-chan patient, client *docker.Client) {
	for p := range in {
		go func(p patient) {
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

				err := client.RestartContainer(p.cid, restartTimeout)
				if err != nil {
					log.Errorf("Error restarting container: %s", p.cid)
				}
			}
		}(p)
	}
}

// newBackoff creates a new backoff instance
func newBackoff(min, max time.Duration) *backoff.Backoff {
	b := backoff.Backoff{
		Min:    min,
		Max:    max,
		Factor: 2,
		Jitter: false,
	}
	return &b
}
