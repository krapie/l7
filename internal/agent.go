package internal

import (
	"log"
	"net/http"

	"github.com/krapie/plumber/internal/loadbalancer"
	"github.com/krapie/plumber/internal/loadbalancer/maglev"
)

type Config struct {
	ServiceDiscoveryMode string
	TargetFilter         string
	MaglevHashKey        string
}

type Agent struct {
	loadBalancer loadbalancer.LoadBalancer
	Config       *Config
}

func NewAgent(config *Config) (*Agent, error) {
	// TODO(krapie): we fix LB configuration maglev for now, but we can make it configurable
	loadBalancer, err := maglev.NewLB(&maglev.Config{
		ServiceDiscoveryMode: config.ServiceDiscoveryMode,
		TargetFilter:         config.TargetFilter,
		MaglevHashKey:        config.MaglevHashKey,
	})
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

	server := &http.Server{
		Addr:      ":80",
		ConnState: maglev.ConnStateEvent,
	}

	log.Printf("[Agent] Starting server on :80")
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
