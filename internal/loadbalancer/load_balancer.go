package loadbalancer

import (
	"net/http"

	"github.com/krapie/plumber/internal/backend"
)

// LoadBalancer is an interface for a load balancer.
type LoadBalancer interface {
	ServeProxy(rw http.ResponseWriter, req *http.Request)

	// TODO(krapie): deprecate methods below
	AddBackend(b *backend.Backend) error
	GetBackends() []*backend.Backend
}
