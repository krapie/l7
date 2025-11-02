package register

import (
	"errors"

	"github.com/krapie/l7/internal/backend/registry"
)

var (
	ErrRegistryNotSet = errors.New("registry not set")
)

const (
	BackendAddedEvent   = "add"
	BackendRemovedEvent = "remove"

	// TODO(krapie): we termporay use this image for testing, but we can make it configurable
	SCHEME = "http"
	IP     = "0.0.0.0"
)

type BackendEvent struct {
	EventType string
	Actor     string
}

type Register interface {
	SetTargetFilter(targetFilter string)
	SetRegistry(registry *registry.BackendRegistry)
	GetEventChannel() chan BackendEvent
	Initialize() error
	Observe()
}
