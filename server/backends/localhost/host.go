package localhost

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/bserdar/watermelon/server"
)

// Backend for localhost
type Backend struct {
	*server.Host
}

func init() {
	server.RegisterBackend("localhost", func(h *server.Host) server.HostBackend {
		return &Backend{Host: h}
	})
}

// NewSession returns a session wrapping host
func (b Backend) NewSession(s server.Session, h *server.Host) (server.HostSession, error) {
	return &Session{Host: b.Host, Session: s}, nil
}

// Session has a pointer to host
type Session struct {
	Host    *server.Host
	Session server.Session
}

// Close closes a session
func (s *Session) Close() {
}

// WriteFile writes a file on the host
func (s *Session) WriteFile(name string, perms os.FileMode, content []byte) (server.CmdErr, error) {
	f, err := os.Create(name)
	if err != nil {
		return server.CmdErrFromErr(server.Localhost, err), nil
	}
	defer f.Close()
	err = f.Chmod(perms)
	if err != nil {
		return server.CmdErrFromErr(server.Localhost, err), nil
	}
	_, err = f.Write(content)
	if err != nil {
		return server.CmdErrFromErr(server.Localhost, err), nil
	}
	return nil, nil
}

// ReadFile reads a local file
func (s *Session) ReadFile(name string) (os.FileInfo, []byte, server.CmdErr, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return fi, nil, server.CmdErrFromErr(server.Localhost, err), nil
	}
	data, err := ioutil.ReadFile(name)
	return fi, data, server.CmdErrFromErr(server.Localhost, err), nil
}

// Run runs cmd
func (s *Session) Run(cmd string, env map[string]string) (server.HostCommandResponse, error) {
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "/bin/sh"
	}
	args := strings.Split(cmd, " ")
	cmd = args[0]
	args = args[1:]
	s.Session.GetLogger(s.Host).Printf("%s %s", cmd, strings.Join(args, " "))
	command := exec.Command(shell, "-s")
	command.Stdin = strings.NewReader(fmt.Sprintf("%s %s", cmd, strings.Join(args, " ")))
	out := bytes.Buffer{}
	err := bytes.Buffer{}
	command.Stdout = &out
	command.Stderr = &err
	e := command.Run()
	s.Session.GetLogger(s.Host).Printf("out: %s err: %s", out.String(), err.String())
	statusCode := 0
	if e != nil {
		if c, ok := e.(*exec.ExitError); ok {
			statusCode = c.ProcessState.ExitCode()
			e = nil
		} else {
			return server.HostCommandResponse{}, e
		}
	}
	return server.HostCommandResponse{Out: out.Bytes(), Err: err.Bytes(), ExitCode: statusCode}, nil
}

// GetFileInfo retrieves file info from a host
func (s *Session) GetFileInfo(file string) (server.FileOwner, os.FileInfo, server.CmdErr, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return server.FileOwner{}, fi, nil, nil
	}
	fo := server.FileOwner{}
	fo.OwnerID = fmt.Sprint(fi.Sys().(*syscall.Stat_t).Uid)
	fo.GroupID = fmt.Sprint(fi.Sys().(*syscall.Stat_t).Gid)
	u, _ := user.LookupId(fo.OwnerID)
	if u != nil {
		fo.OwnerName = u.Username
	}
	g, _ := user.LookupGroupId(fo.GroupID)
	if g != nil {
		fo.GroupName = g.Name
	}
	return fo, fi, nil, nil
}

// MkDir creates dir
func (s *Session) MkDir(path string) (server.CmdErr, error) {
	return server.CmdErrFromErr(server.Localhost, os.MkdirAll(path, 0775)), nil
}

// Chmod changes file mode
func (s *Session) Chmod(path string, mode int) (server.CmdErr, error) {
	return server.CmdErrFromErr(server.Localhost, os.Chmod(path, os.FileMode(mode))), nil

}

// Chown changes file owner
func (s *Session) Chown(path, u, g string) (server.CmdErr, error) {
	if len(u) > 0 {
		us, _ := user.LookupId(u)
		if us == nil {
			us, _ = user.Lookup(u)
		}
		if us != nil {
			uid, err := strconv.Atoi(us.Uid)
			if err == nil {
				err := os.Chown(path, uid, -1)
				if err != nil {
					return server.CmdErrFromErr(server.Localhost, err), nil
				}
			} else {
				return server.CmdErrFromErr(server.Localhost, err), nil
			}
		} else {
			return server.NewCmdErr(server.Localhost, "User not found: %s", u), nil
		}
	}
	if len(g) > 0 {
		gr, _ := user.LookupGroupId(g)
		if gr == nil {
			gr, _ = user.LookupGroup(g)
		}
		if gr != nil {
			gid, err := strconv.Atoi(gr.Gid)
			if err == nil {
				err := os.Chown(path, -1, gid)
				if err != nil {
					return server.CmdErrFromErr(server.Localhost, err), nil
				}
			} else {
				return server.CmdErrFromErr(server.Localhost, err), nil
			}
		} else {
			return server.NewCmdErr(server.Localhost, "Group not found: %s", g), nil
		}
	}
	return nil, nil
}
