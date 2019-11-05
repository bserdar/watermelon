package session

import (
	"fmt"
	"sync"

	jptr "github.com/dustin/go-jsonpointer"
	log "github.com/sirupsen/logrus"

	"github.com/bserdar/watermelon/server"
)

// Session keeps everything related to one running session
type Session struct {
	sync.RWMutex

	ID        string
	Inv       server.InternalInventory
	Modules   server.ModuleMgr
	Log       server.Logging
	LogStdout bool
	Config    interface{}
}

var sessionCtr = 0

// Factory creates sessions
func Factory() server.Session {
	sessionCtr++
	return &Session{ID: fmt.Sprintf("s-%d", sessionCtr)}
}

// GetID returns session ID
func (s *Session) GetID() string { return s.ID }

// GetInv returns the inventory
func (s *Session) GetInv() server.InternalInventory { return s.Inv }

// SetInv sets the inventory
func (s *Session) SetInv(i server.InternalInventory) { s.Inv = i }

// GetModules returns the module mgr
func (s *Session) GetModules() server.ModuleMgr { return s.Modules }

// SetModules returns the module mgr
func (s *Session) SetModules(m server.ModuleMgr) { s.Modules = m }

// SetLogStdout sets whether to log to stdout
func (s *Session) SetLogStdout(z bool) { s.LogStdout = z }

// SetLog sets logger
func (s *Session) SetLog(l server.Logging) { s.Log = l }

// GetConfig returns the configuration
func (s *Session) GetConfig() interface{} { return s.Config }

// SetConfig sets the configuration
func (s *Session) SetConfig(i interface{}) { s.Config = i }

// Close a session
func (s *Session) Close() {
	server.Sessions.Lock()
	defer server.Sessions.Unlock()
	s.Modules.Close()
	delete(server.Sessions.Sessions, s.ID)
}

// GetLogger returns a logger for the host
func (s *Session) GetLogger(host *server.Host) server.Logger {
	return s.Log.New(host, s.LogStdout)
}

// GetHost returns a host from the session
func (s *Session) GetHost(hostID string) (*server.Host, error) {
	if hostID == server.LocalhostID {
		return server.Localhost, nil
	}
	host, err := s.Inv.GetHost([]string{hostID})
	if err != nil {
		return nil, err
	}
	if len(host) != 1 {
		return nil, fmt.Errorf("Invalid host: %s", hostID)
	}
	return host[0], nil
}

// GetCfg returns a path from cfg, either seen by host, or global config
func (s *Session) GetCfg(hostId, path string) interface{} {
	var cfg interface{}

	log.Debugf("GetCfg with host=%s path=%s", hostId, path)
	usingHostCfg := false
	if len(hostId) > 0 {
		hosts, err := s.Inv.GetHost([]string{hostId})
		if err != nil {
			panic(hostId)
		}
		if len(hosts) == 1 {
			cfg = hosts[0].Configuration
			usingHostCfg = true
			log.Debugf("Using host config: %v", hosts[0].Configuration)
		}
	}

	if cfg == nil {
		cfg = s.Config
		log.Debugf("Using global config: %v", cfg)
	}

	if cfg == nil {
		return nil
	}

	ret := jptr.Get(cfg.(map[string]interface{}), path)
	if ret == nil {

		if usingHostCfg {
			cfg = s.Config
			ret = jptr.Get(cfg.(map[string]interface{}), path)
		}
	}
	if ret == nil {
		return nil
	}
	return ret
}
