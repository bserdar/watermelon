package client

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bserdar/watermelon/server/pb"
)

// Host encapsulates a single host
type Host struct {
	S  *Session
	ID string
}

// LocalhostID is the localhost
const LocalhostID = "localhost"

func (h Host) String() string { return string(h.ID) }

// HasLabel returns if the host has the label
func (h Host) HasLabel(label string) bool {
	for _, x := range h.GetInfo().Labels {
		if x == label {
			return true
		}
	}
	return false
}

// MarshalJSON writes out the ID
func (h Host) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.ID)
}

// Logf logs a message
func (h Host) Logf(format string, args ...interface{}) {
	h.S.Logf(h.ID, format, args...)
}

// GetCfgJSON retrieves a configuration item pointed to by path, a
// JSON pointer, and returns the JSON for it. It looks at host first,
// and if item is not found there, it looks at global configuration
func (h Host) GetCfgJSON(path string) []byte {
	return h.S.GetHostCfgJSON(h.ID, path)
}

// GetCfg retrieves a configuration ite pointed to by path, a JSON
// pointer, and unmarshals the JSON to out
func (h Host) GetCfg(path string, out interface{}) {
	h.S.GetHostCfg(h.ID, path, out)
}

// Command executes a command on the host
func (h Host) Command(cmd string) CmdResponse { return h.S.Command(h.ID, cmd) }

// CommandMayFail returns error if command fails, instead of panicking
func (h Host) CommandMayFail(cmd string) (CmdResponse, error) { return h.S.CommandMayFail(h.ID, cmd) }

// Commandf executes a command on the host
func (h Host) Commandf(format string, args ...interface{}) CmdResponse {
	return h.S.Commandf(h.ID, format, args...)
}

// ReadFile reads from a remote file
func (h Host) ReadFile(file string) (os.FileInfo, []byte) { return h.S.ReadFile(h.ID, file) }

// WriteFile writes a file to a remote host
func (h Host) WriteFile(file string, perms os.FileMode, data []byte) error {
	return h.S.WriteFile(h.ID, file, perms, data)
}

// WriteFileIfDifferent writes a file to a remote host if writing changes the file
func (h Host) WriteFileIfDifferent(file string, perms os.FileMode, data []byte) (bool, error) {
	return h.S.WriteFileIfDifferent(h.ID, file, perms, data)
}

// WriteFileFromTemplate writes a file to a remote host based on a
// template. TemplateData is marshaled in JSON
func (h Host) WriteFileFromTemplate(file string, perms os.FileMode, template string, templateData interface{}) (bool, error) {
	return h.S.WriteFileFromTemplate(h.ID, file, perms, template, templateData)
}

// WriteFileFromTemplateFile writes a file to a remote host based on a
// template file on localhost. TemplateData is marshaled in JSON
func (h Host) WriteFileFromTemplateFile(file string, perms os.FileMode, templateFile string, templateData interface{}) (bool, error) {
	return h.S.WriteFileFromTemplateFile(h.ID, file, perms, templateFile, templateData)
}

// CopyFromLocal copies a file from local to host
func (h Host) CopyFromLocal(fromPath, toPath string) error {
	return h.S.CopyFile(LocalhostID, fromPath, h.ID, toPath)
}

// CopyFromLocalIfDifferent copies a file from local to h if different
func (h Host) CopyFromLocalIfDifferent(fromPath, toPath string) (bool, error) {
	return h.S.CopyFromLocalIfDifferent(fromPath, h.ID, toPath)
}

// WaitHost waits until host becomes available
func (h Host) WaitHost(timeout time.Duration) error {
	return h.S.WaitHost(h.ID, timeout)
}

// GetFileInfo retrieves file information
func (h Host) GetFileInfo(path string) (os.FileInfo, FileOwner) { return h.S.GetFileInfo(h.ID, path) }

// Exists returns true if path exists
func (h Host) Exists(path string) bool { return h.S.Exists(h.ID, path) }

// Mkdir creates a dir. Returns OS error msg
func (h Host) Mkdir(path string) error { return h.S.Mkdir(h.ID, path) }

// Chmod runs chmod on host. Returns OS error msg
func (h Host) Chmod(path string, mode int) error { return h.S.Chmod(h.ID, path, mode) }

// Chown runs chown on host. Returns OS error msg
func (h Host) Chown(path string, user, group string) error { return h.S.Chown(h.ID, path, user, group) }

// Ensure a file has certain attributes. Returns true if things changed
func (h Host) Ensure(path string, req Ensure) bool { return h.S.Ensure(h.ID, path, req) }

// GetInfo returns host info. Panics on invalid host
func (h Host) GetInfo() pb.HostInfo {
	info := h.S.GetHostInfo([]string{h.ID})
	if len(info) != 1 || info[0] == nil {
		panic(fmt.Sprintf("Invalid host: %s", h.ID))
	}
	return *info[0]
}
