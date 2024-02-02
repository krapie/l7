package loadbalancer

import (
	"log"
	"net/http"

	"github.com/krapie/plumber/internal/backend"
)

type RoundRobinLB struct {
	backends []*backend.Backend
	index    int
}

func NewRoundRobinLB() (*RoundRobinLB, error) {
	return &RoundRobinLB{
		backends: []*backend.Backend{},
		index:    0,
	}, nil
}

func (lb *RoundRobinLB) AddBackend(b *backend.Backend) error {
	lb.backends = append(lb.backends, b)
	return nil
}

func (lb *RoundRobinLB) ServeProxy(rw http.ResponseWriter, req *http.Request) {
	if len(lb.backends) == 0 {
		panic("No backends")
	}

	log.Printf("[LoadBalancer] Serving request to backend %s", lb.backends[lb.index].Addr)

	lb.index = (lb.index + 1) % len(lb.backends)
	lb.backends[lb.index].Serve(rw, req)
}
