package round_robin

import (
	"log"
	"net/http"
	"sync/atomic"

	"github.com/krapie/plumber/internal/backend"
	"github.com/krapie/plumber/internal/backend/health"
	"github.com/krapie/plumber/internal/backend/register"
	"github.com/krapie/plumber/internal/backend/register/docker"
	"github.com/krapie/plumber/internal/backend/registry"
)

type RoundRobinLB struct {
	backendRegistry *registry.BackendRegistry
	backendRegister register.Register

	healthChecker *health.Checker

	index int64
}

func NewLB(targetBackendImage string) (*RoundRobinLB, error) {
	backendRegistry := registry.NewRegistry()
	backendRegister, err := docker.NewRegister()
	if err != nil {
		return nil, err
	}

	backendRegister.SetTarget(targetBackendImage)
	backendRegister.SetRegistry(backendRegistry)
	err = backendRegister.Initialize()
	if err != nil {
		return nil, err
	}

	backendRegister.Observe()
	log.Printf("[LoadBalancer] Running backend register")

	healthChecker := health.NewHealthChecker(backendRegistry, 2)
	healthChecker.Run()
	log.Printf("[LoadBalancer] Running health check")

	return &RoundRobinLB{
		backendRegistry: backendRegistry,
		backendRegister: backendRegister,

		healthChecker: healthChecker,

		index: 0,
	}, nil
}

func (lb *RoundRobinLB) AddBackend(b *backend.Backend) error {
	err := lb.backendRegistry.AddBackend(b.ID, b.Addr.String())
	if err != nil {
		return err
	}

	return nil
}

func (lb *RoundRobinLB) GetBackends() []*backend.Backend {
	return lb.backendRegistry.GetBackends()
}

// ServeProxy serves the request to the next backend in the list
// keep in mind that this function and its sub functions need to be thread safe
func (lb *RoundRobinLB) ServeProxy(rw http.ResponseWriter, req *http.Request) {
	if b := lb.getNextBackend(); b != nil {
		log.Printf("[LoadBalancer] Serving request to backend %s", b.Addr.String())
		b.Serve(rw, req)
		return
	}

	http.Error(rw, "No backends available", http.StatusServiceUnavailable)
}

func (lb *RoundRobinLB) getNextBackend() *backend.Backend {
	for i := 0; i < lb.backendRegistry.Len(); i++ {
		index := lb.getNextIndex()

		b, err := lb.backendRegistry.GetBackendByIndex(index)
		if err != nil {
			return nil
		}

		if b.IsAlive() {
			return b
		}
	}

	return nil
}

func (lb *RoundRobinLB) getNextIndex() int64 {
	index := atomic.AddInt64(&lb.index, int64(1)) % int64(lb.backendRegistry.Len())
	atomic.StoreInt64(&lb.index, index)

	return index
}
