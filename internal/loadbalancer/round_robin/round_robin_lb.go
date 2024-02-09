package round_robin

import (
	"log"
	"net/http"
	"sync/atomic"

	"github.com/krapie/plumber/internal/backend"
	"github.com/krapie/plumber/internal/backend/health"
	"github.com/krapie/plumber/internal/backend/register"
	"github.com/krapie/plumber/internal/backend/register/docker"
	"github.com/krapie/plumber/internal/backend/register/k8s"
	"github.com/krapie/plumber/internal/backend/registry"
	"github.com/krapie/plumber/internal/loadbalancer"
)

type RoundRobinLB struct {
	backendRegistry *registry.BackendRegistry
	backendRegister register.Register

	healthChecker *health.Checker

	index int64
}

func NewLB(serviceDiscoveryMode, targetBackendImage string) (*RoundRobinLB, error) {
	backendRegistry := registry.NewRegistry()

	var backendRegister register.Register
	var err error
	if serviceDiscoveryMode == loadbalancer.DiscoveryModeK8s {
		backendRegister, err = k8s.NewRegister()
		if err != nil {
			return nil, err
		}
	} else {
		backendRegister, err = docker.NewRegister()
		if err != nil {
			return nil, err
		}
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
