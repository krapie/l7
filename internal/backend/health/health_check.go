package health

import (
	"net"
	"net/url"
	"time"

	"github.com/krapie/plumber/internal/backend"
	"github.com/krapie/plumber/internal/backend/register"
	"github.com/krapie/plumber/internal/backend/registry"
)

const TCP = "tcp"

type Checker struct {
	backendRegistry *registry.BackendRegistry
	backendRegister register.Register

	interval int
}

func NewHealthChecker(registry *registry.BackendRegistry, register register.Register, interval int) *Checker {
	return &Checker{
		backendRegistry: registry,
		backendRegister: register,

		interval: interval,
	}
}

func (c *Checker) Run() {
	go c.healthCheck()
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
	for _, b := range c.backendRegistry.GetBackends() {
		isAlive := c.checkTCPConnection(b.Addr, c.interval)
		if isAlive && !b.IsAlive() {
			b.SetAlive(backend.ALIVE_UP)
			c.backendRegister.GetEventChannel() <- register.BackendEvent{
				EventType: register.BackendAddedEvent,
				Actor:     b.ID,
			}
		} else if !isAlive {
			b.SetAlive(backend.ALIVE_DOWN)
			c.backendRegister.GetEventChannel() <- register.BackendEvent{
				EventType: register.BackendRemovedEvent,
				Actor:     b.ID,
			}
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
