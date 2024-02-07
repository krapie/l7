package maglev

import (
	"log"
	"net/http"

	"github.com/krapie/plumber/internal/backend"
)

type MaglevLB struct {
	backends            map[string]*backend.Backend
	lookupTable         *Maglev
	healthCheckInterval int
}

func NewLB() (*MaglevLB, error) {
	lookupTable, err := NewMaglev([]string{}, 65537)
	if err != nil {
		return nil, err
	}

	return &MaglevLB{
		backends:            make(map[string]*backend.Backend),
		lookupTable:         lookupTable,
		healthCheckInterval: 2,
	}, nil
}

func (lb *MaglevLB) AddBackend(b *backend.Backend) error {
	lb.backends[b.Addr.String()] = b

	err := lb.lookupTable.Add(b.Addr.String())
	if err != nil {
		return err
	}

	return nil
}

func (lb *MaglevLB) GetBackends() []*backend.Backend {
	var backends []*backend.Backend
	for _, b := range lb.backends {
		backends = append(backends, b)
	}

	return backends
}

// ServeProxy serves the request to the next backend in the list
// keep in mind that this function and its sub functions need to be thread safe
func (lb *MaglevLB) ServeProxy(rw http.ResponseWriter, req *http.Request) {
	// TODO(krapie): Move key extraction from http request header to separate system
	key := req.Header.Get("X-Shard-Key")
	backendKey, err := lb.lookupTable.Get(key)
	if err != nil {
		http.Error(rw, "Error getting backend", http.StatusInternalServerError)
		return
	}

	if b := lb.backends[backendKey]; b != nil {
		log.Printf("[LoadBalancer] Serving request to backend %s", b.Addr.String())
		b.Serve(rw, req)
		return
	}

	http.Error(rw, "No backends available", http.StatusServiceUnavailable)
}
