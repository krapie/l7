package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Backend struct {
	Addr string

	proxy *httputil.ReverseProxy
}

func NewDefaultBackend(addr string) (*Backend, error) {
	parsedAddr, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	return &Backend{
		Addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(parsedAddr),
	}, nil
}

func (b *Backend) Serve(rw http.ResponseWriter, req *http.Request) {
	b.proxy.ServeHTTP(rw, req)
}
