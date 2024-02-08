package loadbalancer

import (
	"net/http"
)

// LoadBalancer is an interface for a load balancer.
type LoadBalancer interface {
	ServeProxy(rw http.ResponseWriter, req *http.Request)
}
