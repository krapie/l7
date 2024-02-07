package internal

import (
	"log"
	"net/http"

	"github.com/krapie/plumber/internal/backend"
	"github.com/krapie/plumber/internal/loadbalancer"
	"github.com/krapie/plumber/internal/loadbalancer/maglev"
)

type Agent struct {
	loadBalancer loadbalancer.LoadBalancer

	serviceDiscovery bool
}

func NewAgent(serviceDiscovery bool, targetBackendImage string) (*Agent, error) {
	// TODO(krapie): we fix LB configuration maglev for now, but we can make it configurable
	loadBalancer, err := maglev.NewLB(targetBackendImage)
	if err != nil {
		return nil, err
	}

	return &Agent{
		loadBalancer: loadBalancer,

		serviceDiscovery: serviceDiscovery,
	}, nil
}

func (s *Agent) Run(backendAddresses []string) error {
	if !s.serviceDiscovery {
		err := s.addBackends(backendAddresses)
		if err != nil {
			return err
		}
	}

	http.HandleFunc("/", s.loadBalancer.ServeProxy)

	// TODO(krapie): temporary specify yorkie related path because http.HandleFunc only support exact match
	http.HandleFunc("/yorkie.v1.YorkieService/ActivateClient", s.loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/DeactivateClient", s.loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/AttachDocument", s.loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/DetachDocument", s.loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/RemoveDocument", s.loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/PushPullChanges", s.loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/WatchDocument", s.loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/Broadcast", s.loadBalancer.ServeProxy)

	log.Printf("[Agent] Starting server on :80")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Agent) addBackends(backendAddresses []string) error {
	for _, addr := range backendAddresses {
		b, err := backend.NewDefaultBackend(addr, addr)
		if err != nil {
			return err
		}

		err = s.loadBalancer.AddBackend(b)
		if err != nil {
			return err
		}
	}

	return nil
}
