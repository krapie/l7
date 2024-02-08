package register

import "github.com/krapie/plumber/internal/backend/registry"

const (
	BackendAddedEvent   = "add"
	BackendRemovedEvent = "remove"
)

type BackendEvent struct {
	EventType string
	Actor     string
}

type Register interface {
	SetTarget(target string)
	SetRegistry(registry *registry.BackendRegistry)
	GetEventChannel() chan BackendEvent
	Observe()
}
