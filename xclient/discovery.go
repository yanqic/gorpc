package xclient

import (
	"errors"
	"sync"
	"math"
	"math/rand"
	"time"
)
type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect // Robbin algorithm
)

type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error) 
}

type MultiServerDiscovery struct {
	r *rand.Rand // generate random number
	mu sync.RWMutex
	servers []string
	index int
}

func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	d := &MultiServerDiscovery{
		servers: servers,
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServerDiscovery)(nil)

func (d *MultiServerDiscovery) Refresh() error {
	return nil
}

func (d *MultiServerDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

func (d *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no avaliable servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index % n] // servers could be update, so mode n to ensure safety
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "",errors.New("rpc discovery: not supported select mode")
	}
}

func (d *MultiServerDiscovery) GetAll() ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	servers := make([]string,len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
