package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

const (
	ALIVE_UP   = true
	ALIVE_DOWN = false
)

type Backend struct {
	ID    string
	Addr  *url.URL
	Alive bool

	mutex sync.RWMutex
	proxy *httputil.ReverseProxy
}

func NewDefaultBackend(ID, addr string) (*Backend, error) {
	parsedAddr, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedAddr)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		http.Error(rw, "Error occurred while processing request", http.StatusBadGateway)
	}

	return &Backend{
		ID:    ID,
		Addr:  parsedAddr,
		Alive: true,

		mutex: sync.RWMutex{},
		proxy: proxy,
	}, nil
}

func (b *Backend) Serve(rw http.ResponseWriter, req *http.Request) {
	b.proxy.ServeHTTP(rw, req)
}

func (b *Backend) SetAlive(alive bool) {
	b.mutex.Lock()
	b.Alive = alive
	b.mutex.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mutex.RLock()
	alive := b.Alive
	b.mutex.RUnlock()

	return alive
}
