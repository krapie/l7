package registry

import (
	"errors"
	"sync/atomic"

	"github.com/krapie/plumber/internal/backend"
)

var (
	ErrIndexOutOfRange      = errors.New("index out of range")
	ErrBackendAlreadyExists = errors.New("backend already exists")
)

type BackendRegistry struct {
	Registry atomic.Value
}

func NewRegistry() *BackendRegistry {
	backendRegistry := atomic.Value{}
	backendRegistry.Store([]*backend.Backend{})

	return &BackendRegistry{
		Registry: backendRegistry,
	}
}

func (s *BackendRegistry) GetBackends() []*backend.Backend {
	return s.Registry.Load().([]*backend.Backend)
}

func (s *BackendRegistry) GetBackendByID(ID string) (*backend.Backend, bool) {
	for _, b := range s.GetBackends() {
		if b.ID == ID {
			return b, true
		}
	}

	return nil, false
}

func (s *BackendRegistry) GetBackendByIndex(index int64) (*backend.Backend, error) {
	backends := s.GetBackends()
	if index < 0 || index >= int64(len(backends)) {
		return nil, ErrIndexOutOfRange
	}

	return backends[index], nil
}

func (s *BackendRegistry) AddBackend(hostname, addr string) error {
	if _, ok := s.GetBackendByID(hostname); ok {
		return ErrBackendAlreadyExists
	}

	b, err := backend.NewDefaultBackend(hostname, addr)
	if err != nil {
		return err
	}

	s.Registry.Store(append(s.GetBackends(), b))

	return nil
}

func (s *BackendRegistry) RemoveBackendByID(ID string) {
	var backends []*backend.Backend
	for _, b := range s.GetBackends() {
		if b.ID != ID {
			backends = append(backends, b)
		}
	}

	s.Registry.Store(backends)
}

func (s *BackendRegistry) Len() int {
	return len(s.GetBackends())
}
