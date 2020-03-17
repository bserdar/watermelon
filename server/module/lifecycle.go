package module

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"

	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/pb"
)

// LifecycleManager deals with managing the external module
// lifecycle. Create a LifecycleManager for the module, and then
// start the module. The lifecycle manager waits for the module to
// connect back to the server.
type LifecycleManager struct {
	sync.RWMutex

	// Modules are looked up from these dirs. Each module is a
	// directory.  Modules use the / separator. Module ookup dirs use
	// the hos specific path separator
	ModuleLookupDirs []string

	// RunModuleScript should run module.w in the given dir. Should
	// return error only if execution fails. First is true if this is the
	// first loading of the module in this run
	RunModuleScript func(first bool, dir string) error

	// LocalModuleFunc is called to see if the module is local. If it
	// is, then this should handle the call to the module, and should
	// return response,true,err. If the module is not local, it should
	// return response,false,nil
	LocalModuleFunc func(session, module, funcName string, data []byte) (server.Response, bool, error)

	modules      map[string]*moduleInfo
	builtModules map[string]struct{}

	// this is used during module load, to store the module name that's being loaded
	nextModuleName string
	connCh         chan error
}

type response struct {
	err error
}

type moduleInfo struct {
	sync.Mutex
	server string
	name   string
	conn   *grpc.ClientConn
	reqCh  chan pb.LifecycleRequest
	respCh chan response

	lastPing time.Time
}

// Pings the module to see if it is still alive
func (m *moduleInfo) Ping() error {
	m.Lock()
	defer m.Unlock()

	m.reqCh <- pb.LifecycleRequest{RequestType: pb.LifecycleRequest_PING}
	rsp := <-m.respCh
	if rsp.err != nil {
		return rsp.err
	}
	m.lastPing = time.Now()
	return nil
}

// Connect is called by the client runtime once the module loads. When
// connected, this will initialize the module information in the
// server, and ping the module. A failed ping will remove the module
// from the module info map. The lifecycle manager must be locked during
// this call
func (mgr *LifecycleManager) Connect(stream pb.Lifecycle_ConnectServer) error {
	log.Debugf("Connect is called, waiting for %s", mgr.nextModuleName)
	if len(mgr.nextModuleName) == 0 || mgr.connCh == nil {
		return fmt.Errorf("Invalid state, unexpected connection")
	}
	// Receive the connect msg
	req, err := stream.Recv()
	if err != nil {
		mgr.connCh <- err
		return err
	}
	// Module is built, tag it
	mgr.builtModules[mgr.nextModuleName] = struct{}{}
	if req.RequestType != pb.LifecycleRequest_CONNECT {
		err = fmt.Errorf("Invalid state, expecting connect")
		mgr.connCh <- err
		return err
	}

	connectMsg := req.GetConnectMsg()
	if _, ok := mgr.modules[mgr.nextModuleName]; ok {
		err = fmt.Errorf("Duplicate module %s", mgr.nextModuleName)
		mgr.connCh <- err
		return err
	}
	log.Debugf("Connect ok")
	mod := &moduleInfo{server: fmt.Sprintf("localhost:%d", connectMsg.Port),
		name:   mgr.nextModuleName,
		reqCh:  make(chan pb.LifecycleRequest),
		respCh: make(chan response)}
	mgr.modules[mod.name] = mod
	mgr.connCh <- nil
	log.Debugf("Connect complete")
	// Loop
	for {
		select {
		case r := <-mod.reqCh:
			switch r.RequestType {
			case pb.LifecycleRequest_PING:
				err := stream.Send(&r)
				if err != nil {
					close(mod.respCh)
					return err
				}
				_, err = stream.Recv()
				mod.respCh <- response{err: err}
			case pb.LifecycleRequest_TERM:
				return nil
			}
		}
	}
}

// LoadModule loads a module and returns it GRPC port
func (mgr *LifecycleManager) LoadModule(ctx context.Context, req *pb.LoadModuleRequest) (*pb.LoadModuleResponse, error) {
	mi, err := mgr.load(req.Module)
	if err != nil {
		return nil, err
	}
	return &pb.LoadModuleResponse{Address: mi.server}, nil
}

// ModuleCall is called by a module to call another module
func (mgr *LifecycleManager) ModuleCall(ctx context.Context, req *pb.ModuleWorkRequest) (*pb.Response, error) {
	ws, err := mgr.SendRequest(req.Req.Session, req.ModuleName, req.Req.FuncName, req.Req.Data)
	if err != nil {
		return nil, err
	}
	return &pb.Response{Success: ws.Success,
		FuncName: ws.FuncName,
		ErrorMsg: ws.ErrorMsg,
		Data:     ws.Data}, nil
}

// Log a message
func (mgr *LifecycleManager) Log(ctx context.Context, req *pb.LogRequest) (*pb.Empty, error) {
	s, h, err := server.GetHostAndSession(req.Session, req.HostId)
	if err != nil {
		return nil, err
	}
	s.GetLogger(h).Print(req.Msg)
	return &pb.Empty{}, nil
}

// Print a message
func (mgr *LifecycleManager) Print(ctx context.Context, req *pb.LogRequest) (*pb.Empty, error) {
	fmt.Printf(req.Msg)
	return &pb.Empty{}, nil
}

// GetArgs returns the args to the program
func (mgr *LifecycleManager) GetArgs(ctx context.Context, req *pb.Session) (*pb.Args, error) {
	session := server.GetSession(req.Session)
	if session == nil {
		return nil, fmt.Errorf("Invalid session %s", req.Session)
	}
	return &pb.Args{Args: session.GetArgs()}, nil
}

// GetCfg returns a configuration item by name
func (mgr *LifecycleManager) GetCfg(ctx context.Context, req *pb.CfgRequest) (*pb.CfgResponse, error) {
	log.Debugf("GetCfg %s", req.Path)
	session := server.GetSession(req.Session)
	if session == nil {
		log.Debugf("Bad session")
		return nil, fmt.Errorf("Invalid session %s", req.Session)
	}
	data := session.GetCfg(req.HostId, req.Path)
	if data == nil {
		log.Debugf("Cfg %s not found", req.Path)
		return &pb.CfgResponse{}, nil
	}
	ret, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	log.Debugf("Returning %s", string(ret))
	return &pb.CfgResponse{Data: ret}, nil
}

// Close and shutdown all modules
func (mgr *LifecycleManager) Close() {
	for len(mgr.modules) > 0 {
		for k := range mgr.modules {
			log.Debugf("Shutting down %s", k)
			mgr.end(k)
			break
		}
	}
}

func (mgr *LifecycleManager) end(name string) {
	mgr.Lock()
	defer mgr.Unlock()
	mod, ok := mgr.modules[name]
	if ok {
		mod.reqCh <- pb.LifecycleRequest{RequestType: pb.LifecycleRequest_TERM}
		if mod.conn != nil {
			mod.conn.Close()
		}
		delete(mgr.modules, name)
	}
}

// NewLifecycleManager returns  a new lifecycle manager  to keep track
// of modules connected to the server
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{modules: make(map[string]*moduleInfo),
		builtModules: make(map[string]struct{})}
}

// SendRequest calls a function in a module with the data, and
// returns the response
func (mgr *LifecycleManager) SendRequest(session, module, funcName string, data []byte) (server.Response, error) {
	logger := log.WithField("mth", "SendRequest")
	logger.Debugf("Calling %s.%s", module, funcName)

	response, ok, err := mgr.LocalModuleFunc(session, module, funcName, data)
	if ok {
		return response, err
	}

	mod, err := mgr.load(module)
	if err != nil {
		return server.Response{}, fmt.Errorf("Module not found: %s", module)
	}
	mgr.Lock()
	if mod.conn == nil {
		logger.Debugf("Dialing %s", mod.server)
		mod.conn, err = grpc.Dial(mod.server, grpc.WithInsecure())
		if err != nil {
			mgr.Unlock()
			return server.Response{}, err
		}
	}
	mod.lastPing = time.Now()
	mgr.Unlock()
	cli := pb.NewRequestProcessorClient(mod.conn)
	logger.Debugf("Calling module %s.%s with %+v", module, funcName, string(data))
	ws, err := cli.Process(context.Background(), &pb.Request{Session: session, FuncName: funcName, Data: data})
	logger.Debugf("Module %s.%s returned: %v %v", module, funcName, ws, err)
	if err != nil {
		return server.Response{}, err
	}
	return server.Response{Success: ws.Success,
		FuncName: ws.FuncName,
		ErrorMsg: ws.ErrorMsg,
		Data:     ws.Data}, nil
}

// load loads a module if it is not loaded
func (mgr *LifecycleManager) load(module string) (*moduleInfo, error) {
	mgr.Lock()
	defer mgr.Unlock()
	mi, ok := mgr.modules[module]
	if ok {
		return mi, nil
	}
	log.Debugf("Loading %s", module)
	// Module not loaded. Now load it
	moduleDir, ok := mgr.SearchModuleDir(module)
	if !ok {
		return nil, fmt.Errorf("Cannot find module: %s", module)
	}
	log.Debugf("Module found under %s", moduleDir)
	// Found the directory containing the module. Run the contents of module.w
	// Setup the lifecycle server to receive connection from this module
	mgr.nextModuleName = module
	mgr.connCh = make(chan error)
	defer func() {
		mgr.nextModuleName = ""
		mgr.connCh = nil
	}()

	go func() {
		first := true
		if _, ok := mgr.builtModules[module]; ok {
			first = false
		}
		// This runs the module.w script under current dir. This will
		// return an error only if execution fails
		err := mgr.RunModuleScript(first, moduleDir)
		if err != nil {
			mgr.connCh <- err
		}
	}()
	log.Debugf("Waiting connect")
	// Wait for the connection from the module
	err := <-mgr.connCh
	log.Debugf("Connection done, err: %v", err)
	// We get a nil or error from this one
	if err != nil {
		return nil, err
	}
	mi, _ = mgr.modules[module]
	return mi, nil
}

// SearchModuleDir finds a module directory from the given name. It
// looks up the search directories to see if such module is under any
// one of them
func (mgr *LifecycleManager) SearchModuleDir(module string) (string, bool) {
	parts := strings.Split(module, "/")
	for _, x := range mgr.ModuleLookupDirs {
		lparts := filepath.SplitList(x)
		modulePath := filepath.Join(append(lparts, parts...)...)
		fi, err := os.Stat(filepath.Join(modulePath, "module.w"))
		if err == nil && !fi.IsDir() {
			return modulePath, true
		}
	}
	return "", false
}

// ExecModule executes the module script and listens to its output
func (mgr *LifecycleManager) ExecModule(name string, args ...string) error {
	log.Debugf("Exec %s %v", name, args)
	cmd := exec.Command(name, args...)

	out := func(in io.Reader, errorf func(format string, args ...interface{})) {
		rd := bufio.NewReader(in)
		for {
			str, err := rd.ReadString('\n')
			if len(str) > 0 {
				errorf("From %s: %s", name, str)
			}
			if err != nil {
				break
			}
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go out(stderr, log.Errorf)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go out(stdout, log.Infof)
	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
