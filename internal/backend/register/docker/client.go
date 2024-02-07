package docker

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/krapie/plumber/internal/backend/registry"
)

// TODO(krapie): we termporay use this image for testing, but we can make it configurable
const (
	SCHEME = "http"
	IP     = "0.0.0.0"
)

var (
	ErrRegistryNotSet = errors.New("registry not set")
)

type Register struct {
	DockerClient    *client.Client
	ServiceRegistry *registry.BackendRegistry
	AdditionalTable registry.Table

	Target string
}

func NewRegister() (*Register, error) {
	dockerCLI, err := client.NewClientWithOpts(client.WithVersion("1.43"))
	if err != nil {
		return nil, err
	}

	return &Register{
		DockerClient: dockerCLI,
	}, nil
}

func (r *Register) SetTarget(target string) {
	r.Target = target
}

func (r *Register) SetRegistry(registry *registry.BackendRegistry) {
	r.ServiceRegistry = registry
}

func (r *Register) SetAdditionalTable(table registry.Table) {
	r.AdditionalTable = table
}

func (r *Register) Initialize() error {
	if r.ServiceRegistry == nil {
		return ErrRegistryNotSet
	}

	containerList, err := r.DockerClient.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("status", "running"),
			filters.Arg("ancestor", r.Target),
		),
	})
	if err != nil {
		return err
	}

	// TODO(krapie): we set the address to localhost for now, but we can make it configurable
	for _, c := range containerList {
		err = r.ServiceRegistry.AddBackend(c.ID, fmt.Sprintf("%s://%s:%d", SCHEME, IP, c.Ports[0].PublicPort))
		if err != nil {
			return err
		}

		if r.AdditionalTable != nil {
			err = r.AdditionalTable.Add(c.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Register) Observe() {
	go r.observe()
}

func (r *Register) observe() {
	// use docker events to observe changes in the container list
	msgCh, errCh := r.DockerClient.Events(context.Background(), types.EventsOptions{
		Filters: filters.NewArgs(
			filters.Arg("type", "container"),
			filters.Arg("image", r.Target),
			filters.Arg("event", "start"),
			filters.Arg("event", "kill"),
		),
	})

	for {
		select {
		case msg := <-msgCh:
			if msg.Action == events.ActionKill {
				r.ServiceRegistry.RemoveBackendByID(msg.Actor.ID)
				if r.AdditionalTable != nil {
					err := r.AdditionalTable.Remove(msg.Actor.ID)
					if err != nil {
						log.Printf("[Register] Error removing backend from additional table: %s", err)
						continue
					}
				}

			} else if msg.Action == events.ActionStart {
				c, err := r.DockerClient.ContainerList(context.Background(), container.ListOptions{
					Filters: filters.NewArgs(
						filters.Arg("id", msg.Actor.ID),
					),
				})
				if err != nil {
					log.Printf("[Register] Error getting container: %s", err)
					continue
				}

				err = r.ServiceRegistry.AddBackend(
					c[0].ID,
					fmt.Sprintf("%s://%s:%d", SCHEME, c[0].Ports[0].IP, c[0].Ports[0].PublicPort),
				)
				if err != nil {
					log.Printf("[Register] Error adding backend: %s", err)
					continue
				}
				if r.AdditionalTable != nil {
					err = r.AdditionalTable.Add(c[0].ID)
					if err != nil {
						log.Printf("[Register] Error adding backend to additional table: %s", err)
						continue
					}
				}
			}
		case err := <-errCh:
			log.Println("[Register] Error:", err)
		}
	}
}
