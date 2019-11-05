package remotelinux

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	scp "github.com/hnakamur/go-scp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/bserdar/watermelon/server"
	sshdial "github.com/bserdar/watermelon/server/ssh"
)

// Backend for a remote linux host
type Backend struct {
	*server.Host
}

func init() {
	server.RegisterBackend("linux", func(h *server.Host) server.HostBackend {
		return &Backend{Host: h}
	})
}

// RemoteSession is a wrapper around ssh session
type RemoteSession struct {
	*ssh.Session

	Client        *sshdial.Client
	Host          *server.Host
	ServerSession server.Session
}

// Close closes a session
func (b *RemoteSession) Close() {
	logger := log.WithField("host", b.Host.ID)
	logger.Debugf("Close session %p", b)
	if b.Session != nil {
		b.Session.Close()
	}
	if b.Client != nil {
		b.Client.Close()
	}
}

// NewSession returns a new session to the host
func (b Backend) NewSession(session server.Session, host *server.Host) (server.HostSession, error) {
	logger := log.WithField("host", host.ID)
	client, err := sshdial.Dial(b.Host)
	if err != nil {
		logger.Warnf("Cannot dial host %s: %s", host.ID, err.Error())
		return nil, err
	}
	ret := &RemoteSession{Client: client, Host: host, ServerSession: session}
	logger.Debugf("Open session: %p", ret)
	return ret, nil
}

// newShellSession returns a new session
func (b *RemoteSession) newShellSession() (*ssh.Session, error) {
	logger := log.WithField("host", b.Host.ID)
	logger.Debugf("Creating new ssh session")
	ss, err := b.Client.SSH.NewSession()
	if err != nil {
		return nil, err
	}
	b.Session = ss
	logger.Debugf("New session: %+v", b.Session)
	return b.Session, nil
}

// // WriteFile writes a remote file via scp
// func (b *RemoteSession) WriteFile(name string, perms os.FileMode, content []byte) (server.CmdErr, error) {
// 	t := time.Now()
// 	tmpFile := fmt.Sprintf("/tmp/wm_%s", uuid.New().String())
// 	fileInfo := scp.NewFileInfo(tmpFile, int64(len(content)), perms&os.ModePerm, t, t)
// 	s := scp.NewSCP(b.Client)
// 	log.Debugf("Writing remote file %s on %s", name, b.Host.ID)
// 	err := s.Send(fileInfo, ioutil.NopCloser(bytes.NewReader(content)), tmpFile)
// 	log.Debugf("Write err: %v", err)
// 	if err != nil {
// 		c := server.CmdErrFromErr(b.Host, err)
// 		log.Errorf("Write file error for %s on %s: %+v", name, b.Host.ID, c)
// 		return c, nil
// 	}
// 	_, errmsg, _, err := b.RunShellCommand(fmt.Sprintf("mv %s %s", tmpFile, name), nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(errmsg) != 0 {
// 		return server.NewCmdErr(b.Host, string(errmsg)), nil
// 	}
// 	return nil, nil
// }

// WriteFile writes a remote file via scp
func (b *RemoteSession) WriteFile(name string, perms os.FileMode, content []byte) (server.CmdErr, error) {
	logger := log.WithField("host", b.Host.ID)
	t := time.Now()
	fileInfo := scp.NewFileInfo(name, int64(len(content)), perms&os.ModePerm, t, t)
	s := scp.NewSCP(b.Client.SSH)
	s = BecomeSCP(b.Host, s)
	logger.Debugf("Writing remote file %s on %s", name, b.Host.ID)
	err := s.Send(fileInfo, ioutil.NopCloser(bytes.NewReader(content)), name)
	logger.Debugf("Write err: %v", err)
	if err != nil {
		c := server.CmdErrFromErr(b.Host, err)
		logger.Errorf("Write file error for %s on %s: %+v", name, b.Host.ID, c)
		return c, nil
	}
	return nil, nil
}

// // ReadFile reads a remote file via scp
// func (b *RemoteSession) ReadFile(name string) (os.FileInfo, []byte, server.CmdErr, error) {
// 	log.Debugf("ReadFile %s", name)
// 	wr := bytes.Buffer{}
// 	tmpFile := fmt.Sprintf("/tmp/wm_%s", uuid.New().String())
// 	_, errmsg, _, err := b.RunShellCommand(fmt.Sprintf("cp %s %s", name, tmpFile), nil)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	_, errmsg, _, err = b.RunShellCommand(fmt.Sprintf("chmod a+r %s", tmpFile), nil)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	if len(errmsg) != 0 {
// 		return nil, nil, server.NewCmdErr(b.Host, string(errmsg)), nil
// 	}

// 	s := scp.NewSCP(b.Client)
// 	log.Debugf("Receive %s", name)
// 	fi, err := s.Receive(tmpFile, &wr)
// 	log.Debugf("Received %d bytes", wr.Len())
// 	b.RunShellCommand(fmt.Sprintf("rm %s", tmpFile), nil)
// 	if err != nil {
// 		log.Debugf("Read error: %s", err.Error())
// 		return nil, nil, server.CmdErrFromErr(b.Host, err), nil
// 	}
// 	return fi, wr.Bytes(), nil, nil
// }

// ReadFile reads a remote file via scp
func (b *RemoteSession) ReadFile(name string) (os.FileInfo, []byte, server.CmdErr, error) {
	logger := log.WithField("host", b.Host.ID)
	logger.Debugf("ReadFile %s", name)
	wr := bytes.Buffer{}
	s := scp.NewSCP(b.Client.SSH)
	s = BecomeSCP(b.Host, s)
	logger.Debugf("Receive %s", name)
	fi, err := s.Receive(name, &wr)
	logger.Debugf("Received %d bytes", wr.Len())
	if err != nil {
		logger.Debugf("Read error: %s", err.Error())
		return nil, nil, server.CmdErrFromErr(b.Host, err), nil
	}
	return fi, wr.Bytes(), nil, nil
}

// Run runs a command on a remote host via ssh
func (b *RemoteSession) Run(cmd string, env map[string]string) (server.HostCommandResponse, error) {
	stdout, stderr, s, err := b.RunShellCommand(cmd, env)
	return server.HostCommandResponse{Out: stdout, Err: stderr, ExitCode: s}, err
}

// RunShellCommand runs a command at a shell on the remote
// host. Returns the output and error
func (b *RemoteSession) RunShellCommand(cmd string, env map[string]string) ([]byte, []byte, int, error) {
	logger := log.WithField("host", b.Host.ID)
	logger.Debugf("Run shell command %s on %s", cmd, b.Host.ID)
	sshSession, err := b.newShellSession()
	if err != nil {
		return nil, nil, 0, err
	}
	defer sshSession.Close()
	for k, v := range env {
		b.Setenv(k, v)
	}
	o := bytes.Buffer{}
	e := bytes.Buffer{}
	b.Stdout = &o
	b.Stderr = &e

	b.ServerSession.GetLogger(b.Host).Printf(cmd)

	cmd = Become(b.Host, cmd)
	logger.Debugf("After become: %s", cmd)
	err = sshSession.Run(cmd)
	logger.Debugf("Ran %s: err: %v", cmd, err)

	out := o.Bytes()
	if len(out) > 0 {
		b.ServerSession.GetLogger(b.Host).Printf("stdout: %s", string(out))
	}
	er := e.Bytes()
	if len(er) > 0 {
		b.ServerSession.GetLogger(b.Host).Printf("stderr: %s", string(er))
	}
	exitStatus := 0
	if err != nil {
		b.ServerSession.GetLogger(b.Host).Printf("exec error: %s", err.Error())
		if c, ok := err.(*ssh.ExitError); ok {
			exitStatus = c.ExitStatus()
			err = nil
		}
	}
	if len(out) > 0 || len(er) > 0 {
		return out, er, exitStatus, nil
	}
	return nil, nil, exitStatus, err
}

// GetFileInfo retrieves file info from a host
func (b *RemoteSession) GetFileInfo(file string) (server.FileOwner, os.FileInfo, server.CmdErr, error) {
	logger := log.WithField("host", b.Host.ID)
	out, e, _, err := b.RunShellCommand(fmt.Sprintf("\\stat -c \"%%s %%f %%u %%U %%g %%G %%X %%Y %%Z %%n\" %s", file), nil)
	if err != nil {
		return server.FileOwner{}, nil, nil, err
	}
	if len(e) > 0 {
		logger.Debugf("Err: %s", string(e))
		if strings.Contains(string(e), "No such file or directory") {
			logger.Debugf("Not found, returning")
			return server.FileOwner{}, nil, nil, nil
		}
	}
	w := server.Words(string(out))
	fo := server.FileOwner{}
	ret := server.CommonFileInfo{}
	if len(w) > 9 {
		x, _ := strconv.Atoi(w[0])
		ret.FileSize = int64(x)
		i, _ := strconv.ParseInt(w[1], 16, 64)
		ret.FileMode = os.FileMode(i)
		fo.OwnerID = w[2]
		fo.OwnerName = w[3]
		fo.GroupID = w[4]
		fo.GroupName = w[5]
		ret.FileName = strings.Join(w[9:], " ")
		ret.FileIsDir = ret.FileMode|os.ModeDir == os.ModeDir
		return fo, ret, nil, nil
	}
	return server.FileOwner{}, nil, server.NewCmdErr(b.Host, "Cannot get file info"), nil
}

// MkDir creates dir
func (b *RemoteSession) MkDir(path string) (server.CmdErr, error) {
	_, _, _, err := b.RunShellCommand(fmt.Sprintf("\\mkdir -p %s", path), nil)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Chmod changes file mode
func (b *RemoteSession) Chmod(path string, mode int) (server.CmdErr, error) {
	_, _, _, err := b.RunShellCommand(fmt.Sprintf("\\chmod 0%o %s", mode, path), nil)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Chown changes file owner
func (b *RemoteSession) Chown(path, u, g string) (server.CmdErr, error) {
	if len(u) > 0 {
		if len(g) > 0 {
			_, e, _, err := b.RunShellCommand(fmt.Sprintf("\\chown %s:%s %s", u, g, path), nil)
			if err != nil {
				return nil, err
			}
			if len(e) > 0 {
				return server.NewCmdErr(b.Host, string(e)), nil
			}
		} else {
			_, e, _, err := b.RunShellCommand(fmt.Sprintf("\\chown %s %s", u, path), nil)
			if err != nil {
				return nil, err
			}
			if len(e) > 0 {
				return server.NewCmdErr(b.Host, string(e)), nil
			}
		}
	} else if len(g) > 0 {
		_, e, _, err := b.RunShellCommand(fmt.Sprintf("\\chown :%s %s", g, path), nil)
		if err != nil {
			return nil, err
		}
		if len(e) > 0 {
			return server.NewCmdErr(b.Host, string(e)), nil
		}
	}
	return nil, nil
}
