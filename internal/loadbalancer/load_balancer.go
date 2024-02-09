package loadbalancer

import (
	"net/http"
)

const (
	DiscoveryModeDocker = "docker"
	DiscoveryModeK8s    = "k8s"
)

// LoadBalancer is an interface for a load balancer.
type LoadBalancer interface {
	ServeProxy(rw http.ResponseWriter, req *http.Request)
}
