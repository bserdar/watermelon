package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/bserdar/watermelon/server/pb"
)

// WorkServer contains the functions defined in this module
type WorkServer struct {
	Functions    Functions
	Server       *grpc.Server
	Listener     net.Listener
	rt           *Runtime
	grpcServices map[string]grpc.ServiceInfo
	grpcServers  map[string]interface{}
	registerGRPC func(*grpc.Server, *Runtime)
}

// Process calls the function
func (w *WorkServer) Process(ctx context.Context, req *pb.Request) (returnResponse *pb.Response, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = fmt.Errorf("Panic: %v", r)
			w.rt.Logf(req.Session, "unknown", "Panic: %v", r)
			returnResponse = nil
		}
	}()

	w.rt.Printf(req.Session, "Running process for %s", req.FuncName)
	sess := w.rt.Session(req.Session)
	// Check if this is one of the declared functions in the module
	if f, ok := w.Functions[req.FuncName]; ok {
		rsp, err := f(sess, req.Data)
		if err != nil {
			return &pb.Response{Success: false,
				FuncName: req.FuncName,
				ErrorMsg: err.Error(),
				Modified: sess.Modified,
				Data:     rsp}, nil
		}
		return &pb.Response{Success: true,
			FuncName: req.FuncName,
			Modified: sess.Modified,
			Data:     rsp}, nil
	}

	// Check if this is one of the registered functions in the module
	segments := strings.SplitN(req.FuncName, ".", 2)
	if len(segments) == 2 {
		// We have to know the prefix to find the implementation
		if f, ok := w.grpcServers[segments[0]]; ok {
			// Is there a function with correct name
			v := reflect.ValueOf(f)
			mth := v.MethodByName(segments[1])
			if mth.IsValid() {
				typ := mth.Type()
				if typ.NumIn() == 2 && typ.NumOut() == 2 && typ.In(1).Kind() == reflect.Ptr {
					dataType := typ.In(1).Elem()
					// Pass session ID as grpc metadata in context
					arg := reflect.New(dataType)
					json.Unmarshal(req.Data, arg.Interface())
					inctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("sid", req.Session))

					result := mth.Call([]reflect.Value{reflect.ValueOf(inctx), arg})
					// result[0] is the return value, result[1] is error
					if !result[1].IsNil() {
						return &pb.Response{Success: false,
							FuncName: req.FuncName,
							ErrorMsg: result[1].Interface().(error).Error()}, nil
					}
					data, _ := json.Marshal(result[0].Interface())
					return &pb.Response{Success: true,
						FuncName: req.FuncName,
						Data:     data}, nil
				}
			}
		}
	}

	return &pb.Response{Success: false,
		FuncName: req.FuncName,
		ErrorMsg: fmt.Sprintf("Not found: %s", req.FuncName)}, nil
}

// Start starts the work server. It doesn't return until the listener is closed
func (w *WorkServer) Start() error {
	// If there are any GRPC services implemented in this module, register them, and get the service info
	// so we can install dispatchers
	if w.registerGRPC != nil {
		w.registerGRPC(w.Server, w.rt)
		w.grpcServices = w.Server.GetServiceInfo()
	}
	pb.RegisterRequestProcessorServer(w.Server, w)
	err := w.Server.Serve(w.Listener)
	return err
}

// Stop the listener
func (w *WorkServer) Stop() {
	if w.Listener != nil {
		w.Listener.Close()
	}
}
