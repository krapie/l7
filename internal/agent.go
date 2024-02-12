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
	httpServer   *http.Server

	Config *Config
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

	http.HandleFunc("/", loadBalancer.ServeProxy)

	// TODO(krapie): temporary specify yorkie related path because http.HandleFunc only support exact match
	http.HandleFunc("/yorkie.v1.YorkieService/ActivateClient", loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/DeactivateClient", loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/AttachDocument", loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/DetachDocument", loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/RemoveDocument", loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/PushPullChanges", loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/WatchDocument", loadBalancer.ServeProxy)
	http.HandleFunc("/yorkie.v1.YorkieService/Broadcast", loadBalancer.ServeProxy)

	httpServer := &http.Server{
		Addr:      ":80",
		ConnState: maglev.ConnStateEvent,
	}

	return &Agent{
		loadBalancer: loadBalancer,
		httpServer:   httpServer,
	}, nil
}

func (s *Agent) Run() error {
	log.Printf("[Agent] Starting server on :80")
	if err := s.httpServer.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
