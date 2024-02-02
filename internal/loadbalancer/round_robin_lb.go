package loadbalancer

import (
	"log"
	"net/http"

	"github.com/krapie/plumber/internal/backend"
)

type RoundRobinLB struct {
	backends            []*backend.Backend
	index               int
	healthCheckInterval int
}

func NewRoundRobinLB() (*RoundRobinLB, error) {
	return &RoundRobinLB{
		backends:            []*backend.Backend{},
		index:               0,
		healthCheckInterval: 2,
	}, nil
}

func (lb *RoundRobinLB) AddBackend(b *backend.Backend) error {
	lb.backends = append(lb.backends, b)
	return nil
}

func (lb *RoundRobinLB) GetBackends() []*backend.Backend {
	return lb.backends
}

// ServeProxy serves the request to the next backend in the list
// keep in mind that this function and its sub functions need to be thread safe
func (lb *RoundRobinLB) ServeProxy(rw http.ResponseWriter, req *http.Request) {
	if b := lb.getNextBackend(); b != nil {
		log.Printf("[LoadBalancer] Serving request to backend %s", lb.backends[lb.index].Addr)
		b.Serve(rw, req)
		return
	}

	http.Error(rw, "No backends available", http.StatusServiceUnavailable)
}

func (lb *RoundRobinLB) getNextBackend() *backend.Backend {
	for i := 0; i < len(lb.backends); i++ {
		lb.index = (lb.index + 1) % len(lb.backends)
		if lb.backends[lb.index].IsAlive() {
			return lb.backends[lb.index]
		}
	}

	return nil
}
