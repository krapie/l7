package health

import (
	"net"
	"net/url"
	"time"

	"github.com/krapie/plumber/internal/backend"
)

const TCP = "tcp"

type Checker struct {
	backends []*backend.Backend
	interval int
}

func NewHealthChecker(interval int) *Checker {
	return &Checker{
		backends: nil,
		interval: interval,
	}
}

func (c *Checker) Run() {
	go c.healthCheck()
}

func (c *Checker) AddBackends(backends []*backend.Backend) {
	c.backends = backends
}

func (c *Checker) healthCheck() {
	t := time.NewTicker(time.Duration(c.interval) * time.Second)
	for {
		select {
		case <-t.C:
			// log.Printf("[Health] Running health check")
			c.checkBackendLiveness()
		}
	}
}

func (c *Checker) checkBackendLiveness() {
	for _, b := range c.backends {
		isAlive := c.checkTCPConnection(b.Addr, c.interval)
		if isAlive {
			b.SetAlive(backend.ALIVE_UP)
		} else {
			b.SetAlive(backend.ALIVE_DOWN)
		}
		// log.Printf("[Health] Backend %s is %v", b.Addr, b.IsAlive())
	}
}

func (c *Checker) checkTCPConnection(addr *url.URL, interval int) bool {
	conn, err := net.DialTimeout(TCP, addr.Host, time.Duration(interval)*time.Second)
	if err != nil {
		return false
	}

	err = conn.Close()
	if err != nil {
		return false
	}

	return true
}
