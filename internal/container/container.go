package container

import (
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/go-kit/kit/log"
	"github.com/jpillora/backoff"
	"github.com/pkg/errors"
)

// Patient represent an unhealthy container
type Patient struct {
	cid     string
	delay   time.Duration
	backoff backoff.Backoff
	attempt int
}

func (p *Patient) incAttempt() {
	p.attempt++
}

// NewDockerClient creates a new Docker Clinet
func NewDockerClient() (*docker.Client, error) {
	if os.Getenv("DOCKER_HOST") == "" {
		client, err := docker.NewClient("unix:///var/run/docker.sock")
		if err != nil {
			return nil, errors.Wrap(err, "Error while get Docker client")
		}
		return client, nil
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "Error while get Docker client from environment")
	}
	return client, nil
}

// Sorter sends unhealthy Container IDs into channel
func Sorter(events <-chan *docker.APIEvents, unhealthyCIDs chan<- string) {
	for event := range events {
		switch event.Status {
		case "health_status: unhealthy":
			unhealthyCIDs <- event.ID
		}
	}
}

// Scheduler determines restart delay for container
func Scheduler(CIDs <-chan string, out chan<- Patient, backoffMin, backoffMax time.Duration, j bool, l log.Logger) {
	patients := make(map[string]Patient)
	var p Patient

	for cid := range CIDs {
		// check is it a new container or not
		if _, seen := patients[cid]; seen {
			l.Log("msg", "I've already seen the container", "cid", cid)
			p = patients[cid]
			p.delay = p.backoff.Duration()
		} else {
			l.Log("msg", "I've never seen the container before", "cid", cid)

			b := newBackoff(backoffMin, backoffMax, j)
			p = Patient{
				cid:     cid,
				backoff: *b,
				delay:   b.Duration(),
			}
		}

		p.incAttempt()
		patients[cid] = p
		l.Log(
			"msg", "Patient scheduled",
			"attempt", p.attempt,
			"delay", p.delay,
			"cid", cid,
		)

		// TODO: Implement delete containers from map

		out <- p
	}
}

// Restarter receives a patient, and wait before restart the patient
func Restarter(in <-chan Patient, client *docker.Client, l log.Logger) {
	for p := range in {
		go func(p Patient) {
			cont, _ := client.InspectContainer(p.cid)
			if cont.State.Health.Status == "unhealthy" {
				l.Log("msg", "Sleeping before restart", "time", p.delay)
				time.Sleep(p.delay)

				l.Log(
					"msg", "Healing patient",
					"function", "restarter",
					"attempt", p.attempt,
					"delay", p.delay,
					"cid", p.cid,
					"container_name", cont.Name,
				)

				err := client.RestartContainer(p.cid, 10)
				if err != nil {
					l.Log("err", "Error restarting container", "cid", p.cid)
				}
			}
		}(p)
	}
}

// PushAlredUnhealhy pushes already running unhealthy containers into unhealthy channel.
func PushAlredUnhealhy(client *docker.Client, cid chan<- string, l log.Logger) {
	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}
	for _, c := range containers {
		cont, err := client.InspectContainer(c.ID)
		if err != nil {
			l.Log("err", "Cannot get container")
		}
		if cont.State.Health.Status == "unhealthy" {
			l.Log("msg", "Healing...", "cid", c.ID)
			err = client.RestartContainer(c.ID, 10)
			if err != nil {
				l.Log("err", "Error restarting container", "cid", c.ID)
			}
			l.Log("msg", "Restart container is OK", "cid", c.ID)
		}
	}
}

// newBackoff creates a new backoff instance
func newBackoff(min, max time.Duration, j bool) *backoff.Backoff {
	b := backoff.Backoff{
		Min:    min,
		Max:    max,
		Factor: 2,
		Jitter: j,
	}
	return &b
}
