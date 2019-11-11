package server

import (
	"os"
)

// HostCommandResponse contains the response of the command
type HostCommandResponse struct {
	Out      []byte
	Err      []byte
	ExitCode int
}

// HostBackend opens a new session to run commands on the host
type HostBackend interface {
	NewSession(Session, *Host) (HostSession, error)
}

// HostSession declares the operations that can run on the host
type HostSession interface {
	WriteFile(name string, perms os.FileMode, content []byte) (CmdErr, error)
	ReadFile(name string) (os.FileInfo, []byte, CmdErr, error)
	Run(cmd string, env map[string]string) (HostCommandResponse, error)
	GetFileInfo(file string) (FileOwner, os.FileInfo, CmdErr, error)
	MkDir(string) (CmdErr, error)
	Chmod(string, int) (CmdErr, error)
	Chown(string, string, string) (CmdErr, error)
	Close()
}

var backends = map[string]func(*Host) HostBackend{}

// RegisterBackend registers a new backend
func RegisterBackend(name string, b func(*Host) HostBackend) {
	backends[name] = b
}

// GetBackend returns the backend
func GetBackend(name string, h *Host) HostBackend {
	b, _ := backends[name]
	if b == nil {
		return nil
	}

	return b(h)
}
