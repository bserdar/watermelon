/*
Package client is the Go runtime for Watermelon modules.

A module is an executable that connects to the Watermelon GRPC
server and accepts commands from the server. The commands are
usually requests to execute a function. To execute these functions,
the module must publish them before connecting the server. This is
done in the module main:

  package main

  import (
    "os"
    "github.com/bserdar/watermelon/client"
  )

  func main() {
    f := client.Functions{}
    f.Add("db.Bootstrap", dbBootstrap)
    f.Add("db.Config", dbConfig)

    f.Add("controller.Bootstrap", controllerBootstrap)

    client.Run(os.Args[1:], f, nil)
  }

The above main declares three functions, assigns them to Go functions
and calls the client runtime with program arguments. The program
arguments will contain the Watermelon server address, and optional
`--log loglevel` argument. The client runtime connect the server,
accepts function call requests and dispatches them. The runtime also
deals with requests made from within this module, such as calling
other modules, or calling the server to get inventory/host
information, or to run a command on a remote host.

In addition to registering functions as above, this runtime allows
publishing functions as part of a GRPC server as well. You can implement
the module as a GRPC server, and publish the server as follows:

  package main

  import (
  	"os"

    grpc "google.golang.org/grpc"

    "github.com/bserdar/watermelon/client"
    "github.com/bserdar/watermelon/modules/pkg/yum"
  )

  func main() {
    yumServer := yum.Server{}
    client.Run(os.Args[1:], nil, func(server *grpc.Server, rt *client.Runtime) {
      // Notify the runtime that there is a GRPC server for "yum"
      rt.RegisterGRPCServer(&yumServer, "yum")
      // Register the server
      yum.RegisterYumServer(server, yumServer)
    })
  }


The published functions are called with a client session, and
arguments. The arguments is a JSON document containing the arguments
to the function. When completed, the function returns a JSON document
containing the response, and an optional error. The function should
either return a non-nil response, or a non-nil error, or both nils.

  func dbBootstrap(session *client.Session, args[]byte) (output []byte,error) {
  }

The Go client runtime has a function wrapper for convenience, so you
can declare functions also as follows:

  // Does not get any arguments, does not return an output or error
  func dbBootstrap(session *client.Session) {
  }

  // Accepts arguments in a InStruct, and returns the output in OutStruct. The
  // wrapper unmarshals input arguments into InStruct, and marshals the output
  // to OutStruct
  func dbBootstrap(session *client.Session,in InStruct) OutStruct {
  }

  // No input/output, but may return error
  func dbBootstrap(client *client.Session) error {
  }


If the service is registered as a GRPC service, the functions are regular
GRPC functions that return ModuleResponse:

  // Yum update GRPC function
  func (s Server) Update(ctx context.Context, req *PackageParams) (*module.Response, error) {
  }

If a module publishes GRPC functions, those functions can be called
from other modules via GRPC, or by calling session.Call. If the module
published functions without GRPC, then only session.Call can be used
to call them.


*/
package client
