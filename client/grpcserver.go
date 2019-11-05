package client

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/metadata"
)

// GRPCServer has the dispatch function that helps implement expose
// functions as GRPC functions. Embed GRPCServer to implement a custom
// GRPC server for a module.
//
//  type MyServer struct {
//     client.GRPCServer
//  }
//
//  func (s MyServer) Function(req *Request) (*Response,errror) {
//     var out Response
//     err:=s.Dispatch(req.SessionID,function,*req,&out)
//     if err!=nil {
//        return nil,err
//     }
//     return &out, nil
//  }
type GRPCServer struct {
	RT *Runtime
}

type grpcServer interface {
	setRuntime(rt *Runtime)
}

// setRuntime sets the runtime of the server
func (s *GRPCServer) setRuntime(rt *Runtime) {
	s.RT = rt
}

// GetSession returns a session with the given ID
func (s *GRPCServer) GetSession(ID string) *Session {
	return s.RT.Session(ID)
}

// Dispatch calls f
func (s GRPCServer) Dispatch(sessionID string, f, input, output interface{}) error {
	session := Session{Rt: s.RT, ID: sessionID}
	var inputData []byte
	var err error
	if input != nil {
		inputData, err = json.Marshal(input)
		if err != nil {
			return err
		}
	}
	out, err := Wrap(f)(&session, inputData)
	if err != nil {
		return err
	}
	if output != nil {
		json.Unmarshal(out, output)
	}
	return nil
}

// SessionFromContext retrieves the session from the GRPC context on
// the receiving and of a GRPC call
func (s GRPCServer) SessionFromContext(ctx context.Context) *Session {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		panic("No session data in grpc context")
	}
	v := md.Get("sid")
	if len(v) != 1 {
		panic("No session data in grpc context")
	}
	return s.RT.Session(v[0])
}
