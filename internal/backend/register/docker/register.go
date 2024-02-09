package docker

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/krapie/plumber/internal/backend/register"
	"github.com/krapie/plumber/internal/backend/registry"
)

type Register struct {
	DockerClient    *client.Client
	ServiceRegistry *registry.BackendRegistry
	EventChannel    chan register.BackendEvent

	TargetFilter string
}

func NewRegister() (*Register, error) {
	dockerCLI, err := client.NewClientWithOpts(client.WithVersion("1.43"))
	if err != nil {
		return nil, err
	}

	return &Register{
		DockerClient: dockerCLI,

		EventChannel: make(chan register.BackendEvent),
	}, nil
}

func (r *Register) GetEventChannel() chan register.BackendEvent {
	return r.EventChannel
}

func (r *Register) SetTargetFilter(targetFilter string) {
	r.TargetFilter = targetFilter
}

func (r *Register) SetRegistry(registry *registry.BackendRegistry) {
	r.ServiceRegistry = registry
}

func (r *Register) Initialize() error {
	if r.ServiceRegistry == nil {
		return register.ErrRegistryNotSet
	}

	containerList, err := r.DockerClient.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("status", "running"),
			filters.Arg("ancestor", r.TargetFilter),
		),
	})
	if err != nil {
		return err
	}

	// TODO(krapie): we set the address to localhost for now, but we can make it configurable
	for _, c := range containerList {
		err = r.ServiceRegistry.AddBackend(
			c.ID,
			fmt.Sprintf("%s://%s:%d", register.SCHEME, register.IP, c.Ports[0].PublicPort),
		)
		if err != nil {
			return err
		}

		r.EventChannel <- register.BackendEvent{
			EventType: register.BackendAddedEvent,
			Actor:     c.ID,
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
			filters.Arg("image", r.TargetFilter),
			filters.Arg("event", "start"),
			filters.Arg("event", "kill"),
		),
	})

	for {
		select {
		case msg := <-msgCh:
			if msg.Action == events.ActionKill {
				r.ServiceRegistry.RemoveBackendByID(msg.Actor.ID)
				r.EventChannel <- register.BackendEvent{
					EventType: register.BackendRemovedEvent,
					Actor:     msg.Actor.ID,
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
					fmt.Sprintf("%s://%s:%d", register.SCHEME, c[0].Ports[0].IP, c[0].Ports[0].PublicPort),
				)
				if err != nil {
					log.Printf("[Register] Error adding backend: %s", err)
					continue
				}
				r.EventChannel <- register.BackendEvent{
					EventType: register.BackendAddedEvent,
					Actor:     c[0].ID,
				}
			}
		case err := <-errCh:
			log.Println("[Register] Error:", err)
		}
	}
}
