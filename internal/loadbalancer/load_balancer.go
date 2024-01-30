package loadbalancer

import (
	"net/http"

	"github.com/krapie/plumber/internal/backend"
)

// LoadBalancer is an interface for a load balancer.
type LoadBalancer interface {
	AddBackend(b *backend.Backend) error
	ServeProxy(rw http.ResponseWriter, req *http.Request)
}
