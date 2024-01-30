package internal

import (
	"net/http"

	"github.com/krapie/plumber/internal/backend"
	"github.com/krapie/plumber/internal/loadbalancer"
)

type Agent struct {
	lb loadbalancer.LoadBalancer
}

func NewAgent() (*Agent, error) {
	lb, err := loadbalancer.NewRoundRobinLB()
	if err != nil {
		return nil, err
	}

	return &Agent{
		lb: lb,
	}, nil
}

func (s *Agent) Run(backendAddresses []string) error {
	for _, addr := range backendAddresses {
		b, err := backend.NewDefaultBackend(addr)
		if err != nil {
			return err
		}

		err = s.lb.AddBackend(b)
		if err != nil {
			return err
		}
	}

	http.HandleFunc("/", s.lb.ServeProxy)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return err
	}

	return nil
}
