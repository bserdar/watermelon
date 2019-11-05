package inventory

import (
	"fmt"
	"sort"
	"sync"

	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/pb"
)

// hostSet contains references to a set of hosts
type hostSet struct {
	id     string
	hosts  []*server.Host
	refcnt int
}

// add hosts to hostset.
func (h *hostSet) add(session server.Session, hosts ...*server.Host) {
	for _, host := range hosts {
		h.hosts = append(h.hosts, host)
	}
	sort.Slice(h.hosts, func(i, j int) bool { return h.hosts[i].ID < h.hosts[j].ID })
}

// sameHosts compares the hosts of two hostsets, and returns true if
// the two has the same hosts.
func (h *hostSet) sameHosts(x *hostSet) bool {
	if len(h.hosts) == len(x.hosts) {
		for i, j := range h.hosts {
			if j.ID != x.hosts[i].ID {
				return false
			}
		}
		return true
	}
	return false
}

func (h *hostSet) find(id string) *server.Host {
	for _, x := range h.hosts {
		if x.ID == id {
			return x
		}
	}
	return nil
}

// InvServer serves the inventory through named hostsets
type InvServer struct {
	sync.RWMutex

	AllHosts *hostSet
	Sets     map[string]*hostSet
	Session  server.Session
	Cfg      server.InventoryConfiguration
}

// NewInvServer creates a new inventory server instance
func NewInvServer(hosts []*server.Host, cfg server.InventoryConfiguration, session server.Session) *InvServer {
	ret := &InvServer{AllHosts: &hostSet{id: server.AllHosts},
		Sets:    make(map[string]*hostSet),
		Session: session,
		Cfg:     cfg}
	ret.AllHosts.add(session, hosts...)
	ret.Sets[ret.AllHosts.id] = ret.AllHosts
	return ret
}

// Select selects a subset of an inventory based on the given
// criteria, and returns a new inventory id representing the subset
func (srv *InvServer) Select(from string, selectors ...*pb.Selector) (string, error) {
	srv.Lock()
	defer srv.Unlock()

	inv, ok := srv.Sets[from]
	if !ok {
		return "", fmt.Errorf("Inventory not found: %s", from)
	}
	hosts := make([]*server.Host, 0)
	for _, host := range inv.hosts {
		ok := true
		for _, sel := range selectors {
			if !IsMatch(sel, &host.HostInfo) {
				ok = false
				break
			}
		}
		if ok {
			hosts = append(hosts, host)
		}
	}
	newSet := &hostSet{}
	newSet.add(srv.Session, hosts...)
	s := srv.addSet(newSet)
	s.refcnt++
	return s.id, nil
}

func (srv *InvServer) findSet(set *hostSet) *hostSet {
	for _, x := range srv.Sets {
		if x.sameHosts(set) {
			return x
		}
	}
	return nil
}

func (srv *InvServer) addSet(set *hostSet) *hostSet {
	s := srv.findSet(set)
	if s != nil {
		return s
	}
	set.id = server.NewInventoryID()
	srv.Sets[set.id] = set
	return set
}

// Union takes a union of a set of inventories and returns a new inventory
// containing all the hosts in the combined inventories
func (srv *InvServer) Union(req []string) (string, error) {
	srv.Lock()
	defer srv.Unlock()

	hosts := make([]*server.Host, 0)
	for _, src := range req {
		if sourceSet, ok := srv.Sets[src]; ok {
			hosts = append(hosts, sourceSet.hosts...)
		} else {
			return "", fmt.Errorf("Invalid inventory id: %s", src)
		}
	}
	newSet := &hostSet{}
	newSet.add(srv.Session, hosts...)
	s := srv.addSet(newSet)
	s.refcnt++
	return s.id, nil
}

// Make creates a new inventory containing the given hosts
func (srv *InvServer) Make(req []string) (string, error) {
	srv.Lock()
	defer srv.Unlock()

	hosts := make([]*server.Host, 0)
	for _, hid := range req {
		if hid == server.LocalhostID {
			hosts = append(hosts, server.Localhost)
		} else if host := srv.AllHosts.find(hid); host != nil {
			hosts = append(hosts, host)
		} else {
			return "", fmt.Errorf("Invalid host: %s", hid)
		}
	}
	newSet := &hostSet{}
	newSet.add(srv.Session, hosts...)
	s := srv.addSet(newSet)
	s.refcnt++
	return s.id, nil
}

// Add adds new hosts to an inventory
func (srv *InvServer) Add(to string, ids []string) (string, error) {
	srv.Lock()
	defer srv.Unlock()

	hosts := make([]*server.Host, 0)
	if src, ok := srv.Sets[to]; ok {
		hosts = append(hosts, src.hosts...)
		for _, h := range ids {
			if h == server.LocalhostID {
				hosts = append(hosts, server.Localhost)
			} else if host := srv.AllHosts.find(h); host != nil {
				hosts = append(hosts, host)
			} else {
				return "", fmt.Errorf("Invalid host: %s", h)
			}
		}
	} else {
		return "", fmt.Errorf("Invalid inventory: %s", to)
	}
	newSet := &hostSet{}
	newSet.add(srv.Session, hosts...)
	s := srv.addSet(newSet)
	s.refcnt++
	return s.id, nil
}

// GetHostIDs returns the host ids for all the hosts in an inventory
func (srv *InvServer) GetHostIDs(id string) ([]string, error) {
	srv.RLock()
	defer srv.RUnlock()

	if h, ok := srv.Sets[id]; ok {
		ret := make([]string, 0, len(h.hosts))
		for _, h := range h.hosts {
			ret = append(ret, h.ID)
		}
		return ret, nil
	}
	return nil, fmt.Errorf("Hostset not found: %s", id)
}

// GetHost returns information about some hosts
func (srv *InvServer) GetHost(req []string) ([]*server.Host, error) {
	srv.RLock()
	defer srv.RUnlock()

	ret := make([]*server.Host, 0)
	for _, id := range req {
		if h := srv.AllHosts.find(id); h != nil {
			ret = append(ret, h)
		} else {
			return nil, fmt.Errorf("Host not found: %s", id)
		}
	}
	return ret, nil
}

// GetHostInfo returns information about some hosts
func (srv *InvServer) GetHostInfo(req []string) ([]*pb.HostInfo, error) {
	ret, err := srv.GetHost(req)
	if err != nil {
		return nil, err
	}
	out := make([]*pb.HostInfo, 0, len(ret))
	for _, x := range ret {
		out = append(out, &x.HostInfo)
	}
	return out, nil
}

// Release notifies the server that the inventory id is no longer in use
func (srv *InvServer) Release(id string) {
	srv.Lock()
	defer srv.Unlock()

	set, ok := srv.Sets[id]
	if ok {
		set.refcnt--
		if set.refcnt <= 0 {
			delete(srv.Sets, id)
		}
	}
}
