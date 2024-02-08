package maglev

import (
	"log"
	"net/http"

	"github.com/krapie/plumber/internal/backend/health"
	"github.com/krapie/plumber/internal/backend/register"
	"github.com/krapie/plumber/internal/backend/register/docker"
	"github.com/krapie/plumber/internal/backend/registry"
)

const (
	MinVirtualNodes = 65537
)

type MaglevLB struct {
	backendRegistry *registry.BackendRegistry
	backendRegister register.Register

	lookupTable *Maglev
}

func NewLB(targetBackendImage string) (*MaglevLB, error) {
	lookupTable, err := NewMaglev([]string{}, MinVirtualNodes)
	if err != nil {
		return nil, err
	}

	backendRegistry := registry.NewRegistry()
	backendRegister, err := docker.NewRegister()
	if err != nil {
		return nil, err
	}

	backendRegister.SetTarget(targetBackendImage)
	backendRegister.SetRegistry(backendRegistry)
	backendRegister.SetAdditionalTable(lookupTable)
	err = backendRegister.Initialize()
	if err != nil {
		return nil, err
	}

	backendRegister.Observe()
	log.Printf("[LoadBalancer] Running backend register")

	healthChecker := health.NewHealthChecker(backendRegistry, 2)
	healthChecker.Run()
	log.Printf("[LoadBalancer] Running health check")

	return &MaglevLB{
		backendRegistry: backendRegistry,
		backendRegister: backendRegister,

		lookupTable: lookupTable,
	}, nil
}

// ServeProxy serves the request to the next backend in the list
// keep in mind that this function and its sub functions need to be thread safe
func (lb *MaglevLB) ServeProxy(rw http.ResponseWriter, req *http.Request) {
	// TODO(krapie): Move key extraction from http request header to separate system
	key := req.Header.Get("X-Shard-Key")
	if key == "" {
		key = "default"
	}

	backendKey, err := lb.lookupTable.Get(key)
	if err != nil {
		http.Error(rw, "Error getting backend", http.StatusInternalServerError)
		return
	}

	b, exists := lb.backendRegistry.GetBackendByID(backendKey)
	if !exists {
		http.Error(rw, "No backends available", http.StatusServiceUnavailable)
		return
	}

	b.Serve(rw, req)
}
