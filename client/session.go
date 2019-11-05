package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/bserdar/watermelon/server/pb"
)

// Session represents a client session. All session functions will
// panic if there is an error
type Session struct {
	Rt       *Runtime
	ID       string
	Modified bool
}

// CommandError is an error message from a host
type CommandError struct {
	Host string
	Msg  string
}

// Error returns host: msg
func (c CommandError) Error() string {
	return fmt.Sprintf("%s: %s", c.Host, c.Msg)
}

// Logf logs a message
func (s Session) Logf(hostID string, format string, args ...interface{}) {
	s.Rt.Logf(s.ID, hostID, format, args...)
}

// Printf prints a message
func (s Session) Printf(format string, args ...interface{}) {
	s.Rt.Printf(s.ID, format, args...)
}

// LoadModule loads a module and returns its GRPC location
func (s *Session) LoadModule(module string) (string, error) {
	return s.Rt.LoadModule(module)
}

// ConnectModule loads a module and returns a client connection to it
func (s *Session) ConnectModule(module string) (*grpc.ClientConn, error) {
	return s.Rt.ConnectModule(module)
}

// Context returns a context to be used in GRPC calls. The returned
// contex contains the session ID as metadata
func (s *Session) Context() context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "sid", s.ID)
}

// Call calls another module
func (s *Session) Call(module, function string, data interface{}) *pb.Response {
	s.Logf("Call %s.%s", module, function)
	r, e := s.Rt.Call(s.ID, module, function, data)
	if e != nil {
		panic(e)
	}
	if r.Modified {
		s.Modified = true
	}
	return r
}

// GetCfgJSON retrieves a configuration item pointed to by path, a
// JSON pointer, and returns the JSON for it.
func (s Session) GetCfgJSON(path string) []byte {
	r, e := s.Rt.GetCfgJSON(s.ID, path)
	if e != nil {
		panic(e)
	}
	return r
}

// GetCfg retrieves a configuration ite pointed to by path, a JSON
// pointer, and unmarshals the JSON to out
func (s Session) GetCfg(path string, out interface{}) {
	e := s.Rt.GetCfg(s.ID, path, out)
	if e != nil {
		panic(e)
	}
}

// GetHostCfgJSON retrieves a configuration item pointed to by path, a
// JSON pointer, and returns the JSON for it. It looks at the host
// configuration first, and if not found there, looks at the global
// configuration
func (s Session) GetHostCfgJSON(host, path string) []byte {
	r, e := s.Rt.GetHostCfgJSON(s.ID, host, path)
	if e != nil {
		panic(e)
	}
	return r
}

// GetHostCfg retrieves a configuration ite pointed to by path, a JSON
// pointer, and unmarshals the JSON to out. It looks at the host
// configuration first, and if it is not found there, it looks at the
// global configuration
func (s Session) GetHostCfg(host, path string, out interface{}) {
	e := s.Rt.GetHostCfg(s.ID, host, path, out)
	if e != nil {
		panic(e)
	}
}

// Select a subset of an inventory based on the given criteria, and
// returns a new inventory id representing the subset
func (s Session) Select(from string, what ...Selector) string {
	r, e := s.Rt.Inv.Select(s.ID, from, what...)
	if e != nil {
		panic(e)
	}
	return r
}

// Union combines inventories to build a new host set containing
// all the hosts in the sources
func (s Session) Union(what []string) string {
	r, e := s.Rt.Inv.Union(s.ID, what)
	if e != nil {
		panic(e)
	}
	return r
}

// Make creates a new inventory from the given host IDs
func (s Session) Make(from []string) string {
	r, e := s.Rt.Inv.Make(s.ID, from)
	if e != nil {
		panic(e)
	}
	return r
}

// Add adds new hosts to the given inventory, and returns the new
// inventory ID
func (s Session) Add(to string, hosts []string) string {
	r, e := s.Rt.Inv.Add(s.ID, to, hosts)
	if e != nil {
		panic(e)
	}
	return r
}

// GetHostIDs returns the host IDs included in the inventory
func (s Session) GetHostIDs(invID string) []string {
	r, e := s.Rt.Inv.GetHostIDs(s.ID, invID)
	if e != nil {
		panic(e)
	}
	return r
}

// GetHostInfo returns the host information for the given hosts
func (s Session) GetHostInfo(IDs []string) []*pb.HostInfo {
	r, e := s.Rt.Inv.GetHostInfo(s.ID, IDs)
	if e != nil {
		panic(e)
	}
	return r
}

// GetHosts returns hosts in an inventory
func (s Session) GetHosts(invID string) []*pb.HostInfo {
	r, e := s.Rt.Inv.GetHosts(s.ID, invID)
	if e != nil {
		panic(e)
	}
	return r
}

// Release notifies the server that the inventory is no longer
// needed, and can be freed
func (s Session) Release(id string) {
	s.Rt.Inv.Release(s.ID, id)
}

// Command executes a command on a host
func (s Session) Command(hostID string, cmd string) CmdResponse {
	s.Logf(hostID, "Command  %s", cmd)
	r, e := s.Rt.Rmt.Command(s.ID, hostID, cmd)
	if e != nil {
		panic(e)
	}
	return r
}

// CommandMayFail returns error if command fails, instead of panicking
func (s Session) CommandMayFail(hostID string, cmd string) (CmdResponse, error) {
	return s.Rt.Rmt.Command(s.ID, hostID, cmd)
}

// Commandf executes a command on a host
func (s Session) Commandf(hostID string, format string, args ...interface{}) CmdResponse {
	s.Logf(hostID, format, args...)
	r, e := s.Rt.Rmt.Commandf(s.ID, hostID, format, args...)
	if e != nil {
		panic(e)
	}
	return r
}

// ReadFile reads from a remote file
func (s Session) ReadFile(hostID string, file string) (os.FileInfo, []byte) {
	s.Logf(hostID, "readFile %s", file)
	fi, data, err := s.Rt.Rmt.ReadFile(s.ID, hostID, file)
	if err != nil {
		panic(err)
	}
	return fi, data
}

// WriteFile writes a file to a remote host
func (s *Session) WriteFile(hostID string, file string, perms os.FileMode, data []byte) error {
	s.Logf(hostID, "writeFile %s", file)
	_, c, e := s.Rt.Rmt.WriteFile(s.ID, hostID, file, perms, data, false)
	if e != nil {
		return e
	}
	if c != nil {
		return fmt.Errorf(c.Msg)
	}
	s.Modified = true
	return nil
}

// WriteFileIfDifferent writes a file to a remote host if writing changes the file
func (s *Session) WriteFileIfDifferent(hostID string, file string, perms os.FileMode, data []byte) (bool, error) {
	s.Logf(hostID, "writeFileIfDifferent %s", file)
	mod, c, e := s.Rt.Rmt.WriteFile(s.ID, hostID, file, perms, data, true)
	s.Logf(hostID, "writeFileIfDifferent %s: changed: %v cmderr: %v err: %v", file, mod, c, e)
	if e != nil {
		return false, e
	}
	if c != nil {
		return false, fmt.Errorf(c.Msg)
	}
	if mod {
		s.Modified = true
	}
	return mod, nil
}

// WriteFileFromTemplate writes a file to a remote host based on a
// template. TemplateData is marshaled in JSON
func (s *Session) WriteFileFromTemplate(hostID string, file string, perms os.FileMode, template string, templateData interface{}) (bool, error) {
	s.Logf(hostID, "writeFileFromTemplate %s", file)
	mod, c, e := s.Rt.Rmt.WriteFileFromTemplate(s.ID, hostID, file, perms, template, templateData, true)
	s.Logf(hostID, "writeFileFromTemplate %s: changed: %v cmderr: %v err: %v", file, mod, c, e)

	if e != nil {
		return false, e
	}
	if c != nil {
		return false, fmt.Errorf(c.Msg)
	}
	if mod {
		s.Modified = true
	}
	return mod, nil
}

// WriteFileFromTemplateFile writes a file to a remote host based on a
// template file loaded from localhost. TemplateData is marshaled in JSON
func (s *Session) WriteFileFromTemplateFile(hostID string, file string, perms os.FileMode, templateFile string, templateData interface{}) (bool, error) {
	s.Logf(hostID, "writeFileFromTemplateFile %s", file)
	data, err := ioutil.ReadFile(templateFile)
	if err != nil {
		panic(err)
	}
	t, err := s.WriteFileFromTemplate(hostID, file, perms, string(data), templateData)
	if t {
		s.Modified = true
	}
	return t, err
}

// CopyFile copies a file
func (s *Session) CopyFile(from string, fromPath string, to string, toPath string) error {
	s.Logf(from, "copyFile %s:%s %s:%s", from, fromPath, to, toPath)
	_, err := s.Rt.Rmt.CopyFile(s.ID, from, fromPath, to, toPath, false)
	s.Modified = true
	return err
}

// CopyIfDifferent copies a file if it is different in destination
func (s *Session) CopyIfDifferent(from string, fromPath string, to string, toPath string) (bool, error) {
	s.Logf(from, "copyIfDifferent %s:%s %s:%s", from, fromPath, to, toPath)
	r, err := s.Rt.Rmt.CopyFile(s.ID, from, fromPath, to, toPath, true)
	if r {
		s.Modified = true
	}
	return r, err
}

// CopyFromLocal copies a file from localhost
func (s *Session) CopyFromLocal(fromPath string, to string, toPath string) error {
	s.Logf(LocalhostID, "copyFileFromLocal  %s %s:%s", fromPath, to, toPath)
	fi, err := os.Stat(fromPath)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadFile(fromPath)
	if err != nil {
		panic(err)
	}
	err = s.WriteFile(to, toPath, fi.Mode(), data)
	if err != nil {
		s.Modified = true
	}
	return err
}

// CopyFromLocalIfDifferent copies a file from localhost if different
func (s *Session) CopyFromLocalIfDifferent(fromPath string, to string, toPath string) (bool, error) {
	s.Logf(LocalhostID, "copyFromLocalIfDifferent  %s %s:%s", fromPath, to, toPath)
	fi, err := os.Stat(fromPath)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadFile(fromPath)
	if err != nil {
		panic(err)
	}
	t, err := s.WriteFileIfDifferent(to, toPath, fi.Mode(), data)
	if t {
		s.Modified = true
	}
	return t, err
}

// WaitHost waits until host becomes available
func (s *Session) WaitHost(hostID string, timeout time.Duration) error {
	s.Logf(hostID, "wait")
	return s.Rt.Rmt.WaitHost(s.ID, hostID, timeout)
}

// GetFileInfo retrieves file information
func (s *Session) GetFileInfo(hostID string, path string) (os.FileInfo, FileOwner) {
	s.Logf(hostID, "getFileInfo %s", path)
	f, o, e := s.Rt.Rmt.GetFileInfo(s.ID, hostID, path)
	if e != nil {
		panic(e)
	}
	return f, o
}

// Exists returns true if path is a file or dir
func (s *Session) Exists(hostID string, path string) bool {
	fi, _ := s.GetFileInfo(hostID, path)
	return fi != nil
}

// Mkdir creates a dir. Returns OS error msg
func (s *Session) Mkdir(hostID string, path string) error {
	s.Logf(hostID, "mkdir %s", path)
	r, e := s.Rt.Rmt.Mkdir(s.ID, hostID, path)
	if e != nil {
		panic(e)
	}
	if r == nil {
		return nil
	}
	return CommandError{Host: r.Host, Msg: r.Msg}
}

// Chmod runs chmod on host. Returns OS error msg
func (s *Session) Chmod(hostID string, path string, mode int) error {
	s.Logf(hostID, "chmod %s", path)
	r, e := s.Rt.Rmt.Chmod(s.ID, hostID, path, mode)
	if e != nil {
		panic(e)
	}
	if r == nil {
		return nil
	}
	return CommandError{Host: r.Host, Msg: r.Msg}
}

// Chown runs chown on host. Returns OS error msg
func (s *Session) Chown(hostID string, path string, user, group string) error {
	s.Logf(hostID, "chown %s", path)
	r, e := s.Rt.Rmt.Chown(s.ID, hostID, path, user, group)
	if e != nil {
		panic(e)
	}
	if r == nil {
		return nil
	}
	return CommandError{Host: r.Host, Msg: r.Msg}
}

// Ensure a file has certain attributes. Returns true if things changed
func (s *Session) Ensure(hostID string, path string, req Ensure) bool {
	s.Logf(hostID, "ensure %s", path)
	r, e := s.Rt.Rmt.Ensure(s.ID, hostID, path, req)
	if e != nil {
		panic(e)
	}
	if r {
		s.Modified = true
	}
	return r
}

// Host returns a host object tied to this session
func (s *Session) Host(h string) Host {
	return Host{S: s, ID: h}
}

// ForAll runs f for all hosts in inv. Returns true if everything is fine.
func (s *Session) ForAll(inv string, f func(Host) error) bool {
	var wg sync.WaitGroup
	errLock := sync.Mutex{}
	result := true
	for _, h := range s.GetHostIDs(inv) {
		wg.Add(1)
		host := s.Host(h)
		go func() {
			defer wg.Done()
			err := f(host)
			if err != nil {
				host.Logf("Error: %s", err.Error())
				errLock.Lock()
				result = false
				errLock.Unlock()
			}
		}()
	}
	wg.Wait()
	return result
}

// ForAllSerial runs f for all hosts in inv one by one. Returns true if everything is fine.
func (s *Session) ForAllSerial(inv string, f func(Host) error) bool {
	for _, h := range s.GetHostIDs(inv) {
		host := s.Host(h)
		err := f(host)
		if err != nil {
			host.Logf("Error: %s", err.Error())
			return false
		}
	}
	return true
}

// ForAllSelected selects hosts mathcing the selector from all hosts,
// and call f for each
func (s *Session) ForAllSelected(sel Selector, f func(Host) error) bool {
	return s.ForAll(s.Select(AllHosts, sel), f)
}
