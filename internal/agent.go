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
}

func NewAgent() (*Agent, error) {
	// TODO(krapie): we fix LB configuration maglev for now, but we can make it configurable
	loadBalancer, err := maglev.NewLB()
	if err != nil {
		return nil, err
	}

	return &Agent{
		loadBalancer: loadBalancer,
	}, nil
}

func (s *Agent) Run(serviceDiscovery bool, backendAddresses []string) error {
	if !serviceDiscovery {
		err := s.addBackends(backendAddresses)
		if err != nil {
			return err
		}
	}

	http.HandleFunc("/", s.loadBalancer.ServeProxy)
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
