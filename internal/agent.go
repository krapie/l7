package internal

import (
	"log"
	"net/http"

	"github.com/krapie/plumber/internal/backend"
	"github.com/krapie/plumber/internal/health"
	"github.com/krapie/plumber/internal/loadbalancer"
)

type Agent struct {
	loadBalancer  loadbalancer.LoadBalancer
	healthChecker *health.Checker
}

func NewAgent() (*Agent, error) {
	loadBalancer, err := loadbalancer.NewRoundRobinLB()
	if err != nil {
		return nil, err
	}

	return &Agent{
		loadBalancer:  loadBalancer,
		healthChecker: health.NewHealthChecker(2),
	}, nil
}

func (s *Agent) Run(backendAddresses []string) error {
	err := s.addBackends(backendAddresses)
	if err != nil {
		return err
	}

	s.healthChecker.AddBackends(s.loadBalancer.GetBackends())
	s.healthChecker.Run()
	log.Printf("[Agent] Running health check")

	http.HandleFunc("/", s.loadBalancer.ServeProxy)
	log.Printf("[Agent] Starting server on :80")
	err = http.ListenAndServe(":80", nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Agent) addBackends(backendAddresses []string) error {
	for _, addr := range backendAddresses {
		b, err := backend.NewDefaultBackend(addr)
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
