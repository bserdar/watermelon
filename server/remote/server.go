package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/pb"
)

type srv struct {
}

// Command executes a command on a remote host
func (s srv) Command(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	log.Debugf("Received cmd request: %+v", req)
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		log.Debugf("Cannot get host: %v", err)
		return nil, err
	}
	response, err := h.RunCmd(h.NewCtx(), session, req.Command, nil)
	if err != nil {
		return nil, err
	}
	return &pb.CommandResponse{Stdout: response.Out,
		Stderr:   response.Err,
		ExitCode: int64(response.ExitCode)}, nil
}

// ReadFile reads the contents of a file from a remote host
func (s srv) ReadFile(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}
	fi, data, cerr, err := h.ReadFile(h.NewCtx(), session, req.File)
	if err != nil {
		return nil, err
	}
	ret := &pb.ReadResponse{Data: data,
		Size: int64(len(data))}
	if fi != nil {
		ret.Found = true
		ret.Info = &pb.FileInfo{Name: fi.Name(),
			Size: fi.Size(),
			Mode: int32(fi.Mode()),
			Time: fi.ModTime().Unix(),
			Dir:  fi.IsDir()}
	}
	if cerr != nil {
		ret.Error = cerr.ToPb()
	}
	return ret, nil
}

func (s srv) Template(ctx context.Context, req *pb.TemplateRequest) (*pb.TemplateResponse, error) {
	t, err := template.New("").Parse(req.Template)
	if err != nil {
		log.Debugf("Cannot parse template: %v", err)
		return &pb.TemplateResponse{Error: &pb.CommandError{Msg: fmt.Sprintf("Cannot parse template: %v", err)}}, nil
	}
	var tdata interface{}
	if len(req.Data) > 0 {
		err := json.Unmarshal(req.Data, &tdata)
		if err != nil {
			return nil, err
		}
	}
	newContent := bytes.Buffer{}
	err = t.Execute(&newContent, tdata)
	if err != nil {
		return &pb.TemplateResponse{Error: &pb.CommandError{Msg: fmt.Sprintf("Error running template: %s", err.Error())}}, nil
	}
	return &pb.TemplateResponse{Out: newContent.String()}, nil
}

// WriteFile writes to a remote file
func (s srv) WriteFile(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	logger := log.WithField("writeFile", req.Name)
	logger.Debugf("Begin")
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}

	data := req.GetData()
	// Generate content if it is coming from a template
	if tmpl := req.GetTemplate(); tmpl != nil {
		logger.Debugf("Generating content from template")
		rsp, err := s.Template(ctx, tmpl)
		if err != nil {
			logger.Debugf("Template error: %v", err)
			return nil, err
		}
		if rsp.Error != nil {
			return &pb.WriteResponse{Error: rsp.Error}, nil
		}
		data = []byte(rsp.Out)
	}

	if req.OnlyIfDifferent {
		logger.Debugf("Checking if file changed")
		_, oldData, _, err := h.ReadFile(h.NewCtx(), session, req.Name)
		if err != nil {
			return nil, err
		}
		if bytes.Compare(oldData, data) == 0 {
			logger.Debugf("File did not change")
			return &pb.WriteResponse{}, nil
		}
	}

	logger.Debugf("Writing %d bytes", len(data))
	cerr, err := h.WriteFile(h.NewCtx(), session, req.Name, os.FileMode(req.Perms), data)
	if err != nil {
		logger.Errorf("Write error: %v", err)
		return nil, err
	}
	if cerr != nil {
		logger.Infof("Command error: %v", cerr)
		return &pb.WriteResponse{Error: cerr.ToPb()}, nil
	}
	logger.Debugf("Write complete")
	return &pb.WriteResponse{Modified: true}, nil
}

// CopyFile copies a file
func (s srv) CopyFile(ctx context.Context, req *pb.CopyRequest) (*pb.CopyResponse, error) {
	log.Debugf("CopyFile %+v", req)
	session, fromHost, err := server.GetHostAndSession(req.Session, req.FromHost)
	if err != nil {
		return nil, err
	}
	_, toHost, err := server.GetHostAndSession(req.Session, req.ToHost)
	if err != nil {
		return nil, err
	}
	log.Debugf("Found fromHost, toHost")
	fromCtx := fromHost.NewCtx()
	fi, data, cerr, err := fromHost.ReadFile(fromCtx, session, req.FromPath)
	if err != nil {
		log.Errorf("Cannot read file %s: %v", req.FromPath, err)
		return nil, err
	}
	log.Debugf("Read source file")
	if err != nil {
		return &pb.CopyResponse{Error: cerr.ToPb()}, nil
	}
	if fi == nil {
		s := fmt.Sprintf("File does not exist: %s:%s", req.FromHost, req.FromPath)
		log.Error(s)
		return &pb.CopyResponse{Changed: false, Error: server.NewCmdErr(fromHost, s).ToPb()}, nil
	}

	if req.OnlyIfDifferent {
		log.Debugf("Checking if file changed")
		_, oldData, cerr, err := toHost.ReadFile(toHost.NewCtx(), session, req.ToPath)
		if err != nil {
			log.Debugf("Read dest file failed: %v", err)
			return nil, err
		}
		if cerr != nil {
			log.Debugf("Error reading dest file: %v", cerr)
		} else if bytes.Compare(oldData, data) == 0 {
			log.Debugf("File will not change")
			return &pb.CopyResponse{Changed: false}, nil
		}
	}
	log.Debugf("Writing dest file")

	cerr, err = toHost.WriteFile(toHost.NewCtx(), session, req.ToPath, fi.Mode(), data)
	if err != nil {
		log.Errorf("Cannot write %s: %v", req.ToPath, err)
		return nil, err
	}
	if cerr != nil {
		return &pb.CopyResponse{Error: cerr.ToPb()}, nil
	}
	log.Debugf("Copy done")
	return &pb.CopyResponse{Changed: true}, nil
}

// WaitHost waits until host becomes available
func (s srv) WaitHost(ctx context.Context, req *pb.WaitHostRequest) (*pb.Empty, error) {
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	end := start.Add(time.Duration(req.Timeout))

	c := h.NewCtx()

	for {
		log.Debugf("Waiting for host: %v", h)
		if time.Now().After(end) {
			return &pb.Empty{}, fmt.Errorf("Timeout while waiting for %s to become available", h.ID)
		}

		_, err := c.New(session)
		if err == nil {
			c.Close()
			break
		}
		log.Debugf("Err: %v", err)
		time.Sleep(time.Second * 10)
	}
	return &pb.Empty{}, nil
}

func (s srv) GetFileInfo(ctx context.Context, req *pb.PathRequest) (*pb.GetFileInfoResponse, error) {
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}
	owner, info, cerr, err := h.GetFileInfo(h.NewCtx(), session, req.Path)
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return &pb.GetFileInfoResponse{Error: cerr.ToPb()}, nil
	}
	ret := &pb.GetFileInfoResponse{Owner: &pb.FileOwner{OwnerName: owner.OwnerName,
		OwnerID:   owner.OwnerID,
		GroupName: owner.GroupName,
		GroupID:   owner.GroupID}}
	if info != nil {
		ret.Info = &pb.FileInfo{Name: info.Name(),
			Size: info.Size(),
			Mode: int32(info.Mode()),
			Time: info.ModTime().Unix(),
			Dir:  info.IsDir()}
	}
	return ret, nil
}

func (s srv) Mkdir(ctx context.Context, req *pb.PathRequest) (*pb.OSResponse, error) {
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}
	cerr, err := h.MkDir(h.NewCtx(), session, req.Path)
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return &pb.OSResponse{Error: cerr.ToPb()}, nil
	}
	return &pb.OSResponse{}, nil
}

func (s srv) Chmod(ctx context.Context, req *pb.ChmodRequest) (*pb.OSResponse, error) {
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}
	cerr, err := h.Chmod(h.NewCtx(), session, req.Path, int(req.Mode))
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return &pb.OSResponse{Error: cerr.ToPb()}, nil
	}
	return &pb.OSResponse{}, nil
}

func (s srv) Chown(ctx context.Context, req *pb.ChownRequest) (*pb.OSResponse, error) {
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}
	cerr, err := h.Chown(h.NewCtx(), session, req.Path, req.User, req.Group)
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return &pb.OSResponse{Error: cerr.ToPb()}, nil
	}
	return &pb.OSResponse{}, nil
}

// Ensure a file has certain attributes
func (s srv) Ensure(ctx context.Context, req *pb.EnsureRequest) (*pb.EnsureResponse, error) {
	session, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}
	e := server.FileDesc{}
	if req.SetMode {
		i := int(req.Mode)
		e.Mode = &i
	}

	if req.SetUid {
		i := req.Uid
		e.UID = &i
	}
	if req.SetGid {
		i := req.Gid
		e.GID = &i
	}
	if req.SetUser {
		e.User = &req.User
	}
	if req.SetGroup {
		e.Group = &req.Group
	}
	if req.CheckDir {
		e.Dir = &req.Dir
	}
	b, cerr, err := h.Ensure(h.NewCtx(), session, req.Path, e)
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return &pb.EnsureResponse{Error: cerr.ToPb()}, nil
	}
	return &pb.EnsureResponse{Changed: b}, nil
}

// New returns a new server
func New() pb.RemoteServer {
	return srv{}
}
