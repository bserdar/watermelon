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

	ID         string
	Inv        server.InternalInventory
	Modules    server.ModuleMgr
	Log        server.Logging
	LogStdout  bool
	Config     interface{}
	Extensions map[string]Extension
}

var sessionCtr = 0

// Factory creates sessions
func Factory() server.Session {
	sessionCtr++
	return &Session{ID: fmt.Sprintf("s-%d", sessionCtr), Extensions: map[string]Extension{}}
}

func (s *Session) getExtension(str string) Extension {
	s.Lock()
	extension, ok := s.Extensions[str]
	if !ok {
		extensionFactory, ok := Extensions[str]
		if !ok {
			panic(fmt.Sprintf("Unknown extension: %s", str))
		}
		extension = extensionFactory(s)
		s.Extensions[str] = extension
	}
	s.Unlock()
	return extension
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

// getCfg returns a path from cfg, either seen by host, or global config
func (s *Session) getCfg(hostId, path string) interface{} {
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

// GetCfg returns a path from cfg, either seen by host, or global config
func (s *Session) GetCfg(hostId, path string) interface{} {
	log.Debugf("getCfg %s with host %s", path, hostId)
	ret := s.getCfg(hostId, path)
	if ret == nil {
		log.Debugf("cfg %s not found in %s", path, hostId)
		return ret
	}
	return s.ExpandConfig(ret)
}

func (s *Session) ExpandConfig(in interface{}) interface{} {
	// Expand config recursively
	if m, ok := in.(map[string]interface{}); ok {
		out := make(map[string]interface{})
		for k, v := range m {
			if k == "valueFrom" {
				return s.ExpandRef(v)
			} else {
				out[k] = s.ExpandConfig(v)
			}
		}
		return out
	}
	if a, ok := in.([]interface{}); ok {
		out := make([]interface{}, 0)
		for _, x := range a {
			out = append(out, s.ExpandConfig(x))
		}
		return out
	}
	return in
}

// ExpandRef expands the configuration value using an extension if there is a mathching one
func (s *Session) ExpandRef(in interface{}) interface{} {
	log.Debugf("Expanding cfg reference %v", in)
	m, ok := in.(map[string]interface{})
	if !ok {
		return in
	}
	t, ok := m["type"]
	if !ok {
		return in
	}
	str, ok := t.(string)
	if !ok {
		return in
	}

	extension := s.getExtension(str)

	expander, ok := extension.(ConfigExpander)
	if !ok {
		panic(fmt.Sprintf("Extension %s cannot deal with configuration", str))
	}
	return expander.ExpandConfig(in)
}
