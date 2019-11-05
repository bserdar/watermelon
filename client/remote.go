package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bserdar/watermelon/server/pb"
)

// Remote is the remote runtime implementation for clients
type Remote struct {
	impl pb.RemoteClient
}

// CmdResponse is a command response
type CmdResponse struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// AllOut returns the stdout + stderr
func (c CmdResponse) AllOut() string {
	out := bytes.Buffer{}
	out.Write(c.Stdout)
	out.Write(c.Stderr)
	return out.String()
}

// Out returns the stdout
func (c CmdResponse) Out() string { return string(c.Stdout) }

// Err returns the stderr
func (c CmdResponse) Err() string { return string(c.Stderr) }

// CommonFileInfo is an os.FileInfo
type CommonFileInfo struct {
	FileName    string
	FileSize    int64
	FileMode    os.FileMode
	FileModTime time.Time
	FileIsDir   bool
}

// FileOwner contains owner information
type FileOwner struct {
	OwnerName string
	OwnerID   string
	GroupName string
	GroupID   string
}

// Ensure describes the attributes of a file/directory
type Ensure struct {
	Mode  *int
	UID   string
	GID   string
	User  string
	Group string
	Dir   *bool
}

func (e Ensure) EnsureDir() Ensure {
	t := true
	e.Dir = &t
	return e
}

func (e Ensure) EnsureMode(mode int) Ensure {
	e.Mode = &mode
	return e
}

// Name returns the file name
func (c CommonFileInfo) Name() string { return c.FileName }

// Size returns the file size
func (c CommonFileInfo) Size() int64 { return c.FileSize }

// Mode returns the file mode
func (c CommonFileInfo) Mode() os.FileMode { return c.FileMode }

// ModTime returns the file modification time
func (c CommonFileInfo) ModTime() time.Time { return c.FileModTime }

// IsDir returns true if this is a directory
func (c CommonFileInfo) IsDir() bool { return c.FileIsDir }

// Sys returns nil
func (c CommonFileInfo) Sys() interface{} { return nil }

// Command executes a command on a host
func (r Remote) Command(session string, hostID string, cmd string) (CmdResponse, error) {
	res, err := r.impl.Command(context.Background(), &pb.CommandRequest{Session: session,
		HostId:  hostID,
		Command: cmd})
	if err != nil {
		return CmdResponse{}, err
	}
	return CmdResponse{Stdout: res.Stdout, Stderr: res.Stderr, ExitCode: int(res.ExitCode)}, nil
}

// Commandf executes a command on a host
func (r Remote) Commandf(session string, hostID string, format string, args ...interface{}) (CmdResponse, error) {
	res, err := r.impl.Command(context.Background(), &pb.CommandRequest{Session: session,
		HostId:  hostID,
		Command: fmt.Sprintf(format, args...)})
	if err != nil {
		return CmdResponse{}, err
	}
	return CmdResponse{Stdout: res.Stdout, Stderr: res.Stderr, ExitCode: int(res.ExitCode)}, nil
}

// ReadFile reads from a remote file
func (r Remote) ReadFile(session string, hostID string, file string) (os.FileInfo, []byte, error) {
	res, err := r.impl.ReadFile(context.Background(), &pb.ReadRequest{Session: session,
		HostId: hostID,
		File:   file})
	if err != nil {
		return nil, nil, err
	}
	var fi os.FileInfo
	if res.Info != nil {
		fi = CommonFileInfo{FileName: res.Info.Name,
			FileSize:    res.Info.Size,
			FileMode:    os.FileMode(res.Info.Mode),
			FileModTime: time.Unix(res.Info.Time, 0),
			FileIsDir:   res.Info.Dir}
	}
	return fi, res.Data, nil
}

// WriteFile writes a file to a remote host
func (r Remote) WriteFile(session string, hostID string, file string, perms os.FileMode, data []byte, onlyIfDifferent bool) (bool, *pb.CommandError, error) {
	rsp, err := r.impl.WriteFile(context.Background(), &pb.WriteRequest{Session: session,
		HostId:          hostID,
		Perms:           int64(perms),
		Name:            file,
		Source:          &pb.WriteRequest_Data{Data: data},
		OnlyIfDifferent: onlyIfDifferent})
	if err != nil {
		return false, nil, err
	}
	return rsp.Modified, rsp.Error, nil
}

// WriteFileFromTemplate writes a file to a remote host based on a template. TemplateData is marshaled in JSON
func (r Remote) WriteFileFromTemplate(session string, hostID string, file string, perms os.FileMode, template string, templateData interface{}, onlyIfDifferent bool) (bool, *pb.CommandError, error) {
	var td []byte
	if templateData != nil {
		var err error
		td, err = json.Marshal(templateData)
		if err != nil {
			return false, nil, err
		}
	}
	rsp, err := r.impl.WriteFile(context.Background(), &pb.WriteRequest{Session: session,
		HostId:          hostID,
		Perms:           int64(perms),
		Name:            file,
		Source:          &pb.WriteRequest_Template{Template: &pb.TemplateRequest{Template: template, Data: td}},
		OnlyIfDifferent: onlyIfDifferent})
	if err != nil {
		return false, nil, err
	}
	return rsp.Modified, rsp.Error, nil
}

// CopyFile copies a file
func (r Remote) CopyFile(session string, from string, fromPath string, to string, toPath string, onlyIfDifferent bool) (bool, error) {
	rsp, err := r.impl.CopyFile(context.Background(), &pb.CopyRequest{Session: session,
		FromHost:        from,
		FromPath:        fromPath,
		ToHost:          to,
		ToPath:          toPath,
		OnlyIfDifferent: onlyIfDifferent})
	if err != nil {
		return false, err
	}
	return rsp.Changed, nil
}

// WaitHost waits until host becomes available
func (r Remote) WaitHost(session string, hostID string, timeout time.Duration) error {
	_, err := r.impl.WaitHost(context.Background(), &pb.WaitHostRequest{Session: session,
		HostId:  hostID,
		Timeout: int64(timeout)})
	return err
}

// GetFileInfo retrieves file information
func (r Remote) GetFileInfo(session string, hostID string, path string) (os.FileInfo, FileOwner, error) {
	fi, err := r.impl.GetFileInfo(context.Background(), &pb.PathRequest{Session: session,
		HostId: hostID,
		Path:   path})
	if err != nil {
		return nil, FileOwner{}, err
	}
	var retFi os.FileInfo
	if fi.Info != nil {
		retFi = &CommonFileInfo{FileName: fi.Info.Name,
			FileSize:    fi.Info.Size,
			FileMode:    os.FileMode(fi.Info.Mode),
			FileModTime: time.Unix(fi.Info.Time, 0),
			FileIsDir:   fi.Info.Dir}
	}
	return retFi,
		FileOwner{OwnerName: fi.Owner.OwnerName,
			OwnerID:   fi.Owner.OwnerID,
			GroupName: fi.Owner.GroupName,
			GroupID:   fi.Owner.GroupID}, nil
}

// Mkdir creates a dir
func (r Remote) Mkdir(session string, hostID string, path string) (*pb.CommandError, error) {
	rsp, err := r.impl.Mkdir(context.Background(), &pb.PathRequest{Session: session,
		HostId: hostID,
		Path:   path})
	if err != nil {
		return nil, err
	}

	return rsp.Error, nil
}

// Chmod runs chmod on host
func (r Remote) Chmod(session string, hostID string, path string, mode int) (*pb.CommandError, error) {
	rsp, err := r.impl.Chmod(context.Background(), &pb.ChmodRequest{Session: session,
		HostId: hostID,
		Path:   path,
		Mode:   int32(mode)})
	if err != nil {
		return nil, err
	}
	return rsp.Error, nil
}

// Chown runs chown on host
func (r Remote) Chown(session string, hostID string, path string, user, group string) (*pb.CommandError, error) {
	rsp, err := r.impl.Chown(context.Background(), &pb.ChownRequest{Session: session,
		HostId: hostID,
		Path:   path,
		User:   user,
		Group:  group})
	if err != nil {
		return nil, err
	}
	return rsp.Error, nil
}

// Ensure a file has certain attributes
func (r Remote) Ensure(session string, hostID string, path string, req Ensure) (bool, error) {
	reqpb := &pb.EnsureRequest{Session: session,
		HostId: hostID,
		Path:   path}
	if req.Mode != nil {
		reqpb.SetMode = true
		reqpb.Mode = int32(*req.Mode)
	}
	if len(req.UID) > 0 {
		reqpb.SetUid = true
		reqpb.Uid = req.UID
	}
	if len(req.GID) > 0 {
		reqpb.SetGid = true
		reqpb.Gid = req.GID
	}
	if len(req.User) > 0 {
		reqpb.SetUser = true
		reqpb.User = req.User
	}
	if len(req.Group) > 0 {
		reqpb.SetGroup = true
		reqpb.Group = req.Group
	}
	if req.Dir != nil {
		reqpb.CheckDir = true
		reqpb.Dir = *req.Dir
	}
	rsp, err := r.impl.Ensure(context.Background(), reqpb)
	if err != nil {
		return false, err
	}
	return rsp.Changed, nil
}
