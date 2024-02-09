package maglev

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/krapie/plumber/internal/backend/health"
	"github.com/krapie/plumber/internal/backend/register"
	"github.com/krapie/plumber/internal/backend/register/docker"
	"github.com/krapie/plumber/internal/backend/register/k8s"
	"github.com/krapie/plumber/internal/backend/registry"
	"github.com/krapie/plumber/internal/loadbalancer"
)

const (
	MinVirtualNodes = 65537
)

// TODO(krapie): temporary store all connections in global variable
var connections = make(map[string]net.Conn)

type Connection struct {
	conn      net.Conn
	key       string
	backendID string
}

type MaglevLB struct {
	backendRegistry *registry.BackendRegistry
	backendRegister register.Register

	lookupTable       *Maglev
	streamConnections []*Connection
}

func NewLB(serviceDiscoveryMode, targetBackendImage string) (*MaglevLB, error) {
	lookupTable, err := NewMaglev([]string{}, MinVirtualNodes)
	if err != nil {
		return nil, err
	}

	backendRegistry := registry.NewRegistry()

	var backendRegister register.Register
	if serviceDiscoveryMode == loadbalancer.DiscoveryModeK8s {
		backendRegister, err = k8s.NewRegister()
		if err != nil {
			return nil, err
		}
	} else {
		backendRegister, err = docker.NewRegister()
		if err != nil {
			return nil, err
		}
	}

	lb := &MaglevLB{
		backendRegistry: backendRegistry,
		backendRegister: backendRegister,

		lookupTable:       lookupTable,
		streamConnections: []*Connection{},
	}
	lb.RunWatchEventLoop()

	backendRegister.SetTarget(targetBackendImage)
	backendRegister.SetRegistry(backendRegistry)
	err = backendRegister.Initialize()
	if err != nil {
		return nil, err
	}

	backendRegister.Observe()
	log.Printf("[LoadBalancer] Running backend register")

	healthChecker := health.NewHealthChecker(backendRegistry, 2)
	healthChecker.Run()
	log.Printf("[LoadBalancer] Running health check")

	return lb, nil
}

// ServeProxy serves the request to the next backend in the list
// keep in mind that this function and its sub functions need to be thread safe
func (lb *MaglevLB) ServeProxy(rw http.ResponseWriter, req *http.Request) {
	// TODO(krapie): Move key extraction from http request header to separate system
	key := req.Header.Get("X-Shard-Key")
	if key == "" {
		key = "default"
	}

	backendID, err := lb.lookupTable.Get(key)
	if err != nil {
		http.Error(rw, "Error getting backend", http.StatusInternalServerError)
		return
	}

	if req.URL.Path == "/yorkie.v1.YorkieService/WatchDocument" {
		conn := GetConn(req)
		lb.streamConnections = append(lb.streamConnections, &Connection{
			conn:      conn,
			key:       key,
			backendID: backendID,
		})
	}

	b, exists := lb.backendRegistry.GetBackendByID(backendID)
	if !exists {
		http.Error(rw, "No backends available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("[LoadBalancer] Time: %s URL: %s Backend: %s", time.Now().Format(time.RFC3339), req.URL, b.ID)
	b.Serve(rw, req)
}

func (lb *MaglevLB) RunWatchEventLoop() {
	go lb.watchBackendEvent()
}

func (lb *MaglevLB) watchBackendEvent() {
	eventChannel := lb.backendRegister.GetEventChannel()
	for {
		select {
		case event := <-eventChannel:
			switch event.EventType {
			case register.BackendAddedEvent:
				err := lb.lookupTable.Add(event.Actor)
				if err != nil {
					log.Printf("[LoadBalancer] Error adding backend to lookup table: %s", err)
				}
				lb.closeSplitBrainedConnection()
			case register.BackendRemovedEvent:
				err := lb.lookupTable.Remove(event.Actor)
				if err != nil {
					log.Printf("[LoadBalancer] Error removing backend from lookup table: %s", err)
				}
			}
		}
	}
}

func (lb *MaglevLB) closeSplitBrainedConnection() {
	// check connections and close if recalculated maglev hashed output(backend ID)
	// for the key is different from the connection's backend ID
	for _, c := range lb.streamConnections {
		backendID, err := lb.lookupTable.Get(c.key)
		if err != nil {
			log.Printf("[LoadBalancer] Error getting backend from lookup table: %s", err)
			continue
		}
		if backendID != c.backendID {
			c.conn.Close()
		}
	}
}

func ConnStateEvent(conn net.Conn, event http.ConnState) {
	if event == http.StateActive {
		connections[conn.RemoteAddr().String()] = conn
	} else if event == http.StateHijacked || event == http.StateClosed {
		delete(connections, conn.RemoteAddr().String())
	}
}

func GetConn(r *http.Request) net.Conn {
	return connections[r.RemoteAddr]
}
