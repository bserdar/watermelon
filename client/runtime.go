package client

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"

	grpc "google.golang.org/grpc"

	log "github.com/sirupsen/logrus"

	pb "github.com/bserdar/watermelon/server/pb"
)

// Runtime is the client runtime to be used by module implementations.
type Runtime struct {
	Port       int
	Worker     WorkServer
	Inv        Inventory
	Rmt        Remote
	ClientConn *grpc.ClientConn
	LCClient   pb.LifecycleClient
}

// Flags returns the flagset
func Flags() *flag.FlagSet {
	set := flag.NewFlagSet("", flag.ExitOnError)
	set.String("log", "info", "set loglevel to info or debug")
	return set
}

// Run creates a runtime and runs it
func Run(args []string, funcs Functions, registerGRPCServers func(*grpc.Server, *Runtime)) {

	flagSet := Flags()
	flagSet.Parse(args)

	if flagSet.Lookup("log").Value.String() == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		panic(err)
	}

	allFuncs := make(Functions)
	for k, v := range registeredFunctions {
		allFuncs[k] = v
	}
	for k, v := range funcs {
		allFuncs[k] = v
	}

	rt, err := NewRuntime(flagSet.Arg(0), listener, allFuncs, registerGRPCServers)
	if err != nil {
		panic(err)
	}
	rt.Start()
}

// NewRuntime creates a new runtime with the given server
// connection. It creates a server on the client side using localhost:myport. The
// server must be of the form host:port
func NewRuntime(server string, listener net.Listener, funcs Functions, registerGRPCServers func(*grpc.Server, *Runtime)) (*Runtime, error) {

	myAddr := listener.Addr().String()
	ix := strings.LastIndex(myAddr, ":")
	portStr := myAddr[ix+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	rt := &Runtime{Port: port}
	// Connect the server
	clientConn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	rt.ClientConn = clientConn
	rt.LCClient = pb.NewLifecycleClient(rt.ClientConn)
	rt.Inv = Inventory{impl: pb.NewInventoryClient(rt.ClientConn)}
	rt.Rmt = Remote{impl: pb.NewRemoteClient(rt.ClientConn)}

	// Create a server for calls coming to this module
	rt.Worker = WorkServer{Functions: funcs, rt: rt, grpcServers: make(map[string]interface{})}
	rt.Worker.Server = grpc.NewServer()
	rt.Worker.Listener = listener
	rt.Worker.registerGRPC = registerGRPCServers

	return rt, nil
}

// Session returns a new session
func (rt *Runtime) Session(id string) *Session {
	return &Session{Rt: rt, ID: id}
}

// RegisterGRPCServer registers a GRPC server with the runtime so its
// methods can be exposed as funcNamePrefix.MethodName
func (rt *Runtime) RegisterGRPCServer(server grpcServer, funcNamePrefix string) {
	server.setRuntime(rt)
	rt.Worker.grpcServers[funcNamePrefix] = server
}

// Start the runtime lifecycle and all the servers associated with
// it. This does not return until the lifecycle ends
func (rt *Runtime) Start() error {
	ch := make(chan struct{})
	go func() {
		ch <- struct{}{}
		rt.Worker.Start()
	}()
	<-ch
	err := rt.runLifecycle(rt.Port)
	rt.Worker.Stop()
	return err
}

// Logf writes a log message for the host
func (rt *Runtime) Logf(session string, hostID string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	rt.LCClient.Log(context.Background(), &pb.LogRequest{Session: session, HostId: string(hostID), Msg: msg})
}

// Printf prints a message
func (rt *Runtime) Printf(session string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	rt.LCClient.Print(context.Background(), &pb.LogRequest{Session: session, Msg: msg})
}

func (rt *Runtime) GetArgs(session string) []string {
	ret, _ := rt.LCClient.GetArgs(context.Background(), &pb.Session{Session: session})
	if ret != nil {
		return ret.Args
	}
	return []string{}
}

// Args can be used to pass values to a function
type Args map[string]interface{}

// Call calls another module
func (rt *Runtime) Call(session, module, function string, data interface{}) (*pb.Response, error) {
	var fdata []byte
	if x, ok := data.([]byte); ok {
		fdata = x
	} else if x, ok := data.(string); ok {
		fdata = []byte(x)
	} else {
		fdata, _ = json.Marshal(data)
	}
	response, err := rt.LCClient.ModuleCall(context.Background(), &pb.ModuleWorkRequest{ModuleName: module,
		Req: &pb.Request{Session: session, FuncName: function, Data: fdata}})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetCfgJSON retrieves a configuration item pointed to by path, a
// JSON pointer, and returns the JSON for it.
func (rt *Runtime) GetCfgJSON(session, path string) ([]byte, error) {
	ret, err := rt.LCClient.GetCfg(context.Background(), &pb.CfgRequest{Session: session, Path: path})
	if err != nil {
		return nil, err
	}
	return ret.Data, nil
}

// GetCfg retrieves a configuration ite pointed to by path, a JSON
// pointer, and unmarshals the JSON to out
func (rt *Runtime) GetCfg(session, path string, out interface{}) error {
	x, err := rt.GetCfgJSON(session, path)
	if err != nil {
		return err
	}
	if x == nil {
		return nil
	}
	return json.Unmarshal(x, out)
}

// GetHostCfgJSON retrieves a host-specific configuration item pointed
// to by path, a JSON pointer, and returns the JSON for it. If the
// host does not have the configuration item, this looks at the global
// configuration
func (rt *Runtime) GetHostCfgJSON(session, host, path string) ([]byte, error) {
	ret, err := rt.LCClient.GetCfg(context.Background(), &pb.CfgRequest{Session: session, HostId: host, Path: path})
	if err != nil {
		return nil, err
	}
	return ret.Data, nil
}

// GetHostCfg retrieves a configuration ite pointed to by path, a JSON
// pointer, and unmarshals the JSON to out. It looks at host specific
// configuration first, and then the global configuration
func (rt *Runtime) GetHostCfg(session, host, path string, out interface{}) error {
	x, err := rt.GetHostCfgJSON(session, host, path)
	if err != nil {
		return err
	}
	if x == nil {
		return nil
	}
	return json.Unmarshal(x, out)
}

// runLifecycle connects to the server, responds to pings, and waits
// for the term signal. Once terminated, it returns from this function
// with nil. Any communication error will immediately return
func (rt *Runtime) runLifecycle(serverPort int) error {
	lifecycleCli, err := rt.LCClient.Connect(context.Background())
	if err != nil {
		return err
	}
	err = lifecycleCli.Send(&pb.LifecycleRequest{RequestType: pb.LifecycleRequest_CONNECT,
		Msg: &pb.LifecycleRequest_ConnectMsg{ConnectMsg: &pb.Connect{Port: int32(serverPort)}}})
	if err != nil {
		return err
	}
	for {
		req, err := lifecycleCli.Recv()
		if err != nil {
			return err
		}
		switch req.RequestType {
		case pb.LifecycleRequest_PING:
			err := lifecycleCli.Send(&pb.LifecycleRequest{RequestType: pb.LifecycleRequest_PONG})
			if err != nil {
				return err
			}
		case pb.LifecycleRequest_TERM:
			return nil
		}
	}
}

// LoadModule loads a module and returns its GRPC address
func (rt *Runtime) LoadModule(name string) (string, error) {
	rsp, err := rt.LCClient.LoadModule(context.Background(), &pb.LoadModuleRequest{Module: name})
	if err != nil {
		return "", err
	}
	return rsp.Address, nil
}

// ConnectModule loads a module and returns a client connection to it
func (rt *Runtime) ConnectModule(module string) (*grpc.ClientConn, error) {
	adr, err := rt.LoadModule(module)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(adr, grpc.WithInsecure())
}
