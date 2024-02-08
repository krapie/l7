package internal

import (
	"log"
	"net/http"

	"github.com/krapie/plumber/internal/loadbalancer"
	"github.com/krapie/plumber/internal/loadbalancer/maglev"
)

type Agent struct {
	loadBalancer loadbalancer.LoadBalancer
}

func NewAgent(targetBackendImage string) (*Agent, error) {
	// TODO(krapie): we fix LB configuration maglev for now, but we can make it configurable
	loadBalancer, err := maglev.NewLB(targetBackendImage)
	if err != nil {
		return nil, err
	}

	return &Agent{
		loadBalancer: loadBalancer,
	}, nil
}

func (s *Agent) Run() error {
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
