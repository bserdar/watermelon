package server

import (
	"fmt"
	"sync"
)

var SessionFactory func() Session

// Session interface
type Session interface {
	GetID() string
	// Close a session
	Close()
	// GetLogger returns a logger for the host
	GetLogger(host *Host) Logger
	// GetHost returns a host from the session
	GetHost(hostID string) (*Host, error)
	// GetCfg returns a path from cfg, either seen by host, or global config
	GetCfg(hostId, path string) interface{}

	// GetInv returns inventory
	GetInv() InternalInventory
	SetInv(InternalInventory)
	GetModules() ModuleMgr
	SetModules(ModuleMgr)
	SetLogStdout(bool)
	SetLog(Logging)
	GetConfig() interface{}
	SetConfig(interface{})
}

type sessionRegistry struct {
	sync.RWMutex
	Sessions map[string]Session
}

var Sessions = sessionRegistry{Sessions: map[string]Session{}}

// NewSession creates a new empty session
func NewSession() Session {
	Sessions.Lock()
	defer Sessions.Unlock()

	s := SessionFactory()
	Sessions.Sessions[s.GetID()] = s
	return s
}

// GetSession returns a session by ID. Returns nil if not found
func GetSession(ID string) Session {
	Sessions.RLock()
	ret, _ := Sessions.Sessions[ID]
	Sessions.RUnlock()
	return ret
}

// GetHostFromSession returns a host from the session
func GetHostFromSession(sessionID, hostID string) (*Host, error) {
	session := GetSession(sessionID)
	if session == nil {
		return nil, ErrInvalidSession(sessionID)
	}
	if hostID == string(LocalhostID) {
		return Localhost, nil
	}
	return session.GetHost(hostID)
}

// GetHostAndSession returns a host from the session
func GetHostAndSession(sessionID, hostID string) (Session, *Host, error) {
	session := GetSession(sessionID)
	if session == nil {
		return nil, nil, ErrInvalidSession(sessionID)
	}
	if hostID == string(LocalhostID) {
		return session, Localhost, nil
	}
	h, err := session.GetHost(hostID)
	return session, h, err
}

// ErrInvalidSession returns invalid session error
func ErrInvalidSession(sessionID string) error {
	return fmt.Errorf("Invalid session: %s", sessionID)
}
