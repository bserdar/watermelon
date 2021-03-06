# Watermelon &nbsp; 
[![GoDoc](https://godoc.org/github.com/bserdar/watermelon?status.svg)](https://godoc.org/github.com/bserdar/watermelon)
[![Go Report](https://goreportcard.com/badge/github.com/bserdar/watermelon)](https://goreportcard.com/report/github.com/bserdar/watermelon)



Watermelon is a tool for building infrastructure as code. It lets you
use a regular programming language (as opposed to some specialized
scripting language) to connect to remote machines, run commands, copy
files to manage configuration remotely. Watermelon abstracts the
details of accessing those remote machines. 

Watermelon is written in Go, and currently supports infrastructure
code written in Go. When you write infrastructure code (a *module*),
you import the watermelon client package that provides the runtime for
your code. Using the client, you can write code that looks like this:

```
func OpenPorts(session *client.Session) {
  session.ForAllSelected(client.Has("smtp"), func(host client.Host) error {
    for _, p := range PORTS {
      host.Commandf("firewall-cmd --permanent --add-port=%d/tcp", p)
    }
    host.Command("firewall-cmd --reload")
    return nil
  })
}
```


This piece of code opens some ports on all hosts labeled with `smtp`
using `firewall-cmd`.

Watermelon communicates with the modules you write using gRPC and
JSON. When you run your infrastructure code, watermelon builds your
code, runs it, and communicates with it over gRPC. You don't need to
build your code separately. This lets you work with your infrastructure
code as if it is a script--you can write code, and run it.

From your module, you can call the functions provided by the client
library as well as functions in other modules. The other modules can
be written in languages other than Go, provided there is a client
runtime for that language.

Watermelon is procedural, not declarative. However, if you can write
idempotent infrastructure code, you can use it as a declarative tool
as well. It is very hard to write declarative rules for things such as
install packages, reboot, then continue with other tasks.


## Install

Use `go get`:

```
go get github.com/bserdar/watermelon
```

These are some existing modules you can use:

```
git clone github.com/bserdar/watermelon-modules
```

You will need to pass the directory of these modules to `watermelon` to use
them, or define WM_MODULES environment variable:

```
export WM_MODULES="~/go/src/github.com/bserdar/watermelon-modules"
```

## Running

Using watermelon you run functions exported in modules. 

Run watermelon using:

```
watermelon run --inv inventory.yml --mdir /dir-to-modules/watermelon-modules --mdir /other/module/dir someModule someFunc
```

 * --inv inventory.yml: This will load `inventory.yml` which contains
host definitions and configuration options. The configuration items
can be accessed from the running scripts using JSON pointers.
 * --mdir dir: Each --mdir option will define a directory under which
   modules can be found. Each module has the name of the last
   component of the directory it is in.
 * someModule someFunc: The module and function to run. This will
   build and load the module `someModule` under one of the `--mdir`s,
   and then execute the function `someFunc` in that module.
   


## Inventory and Configurations

The inventory contains remote hosts and host specific or global
configuration options. Inventory file looks like this:

```
# The ssh private key file to use to ssh remote hosts
# If omitted, username and password for each host must
# be provided.
privateKey: /home/user/.ssh/id_rsa

# Passphrase for the private key. If omitted, passphrase will
# be asked via stdin first time it is needed
passphrase: 123abcdef

# Hosts section lists all remote hosts
hosts:
  - 
    # id is the unique host id. This is how scripts
    # identify hosts
    id: host1
    
    # Addresses of this host. This host has two
    # addresses, a primary one and a private one.
    addresses:
      - address: xxx.xxx.xxx.xxx
        name: primary
      - address: xxx.2xxx.xxx.xxx
        name: private

    # ssh parameters for the host. These are hidden from the 
    # scripts. ssh tries all given authentication schemes to 
    # connect a host. Specifying both password and private key
    # will allow ssh using an initial password, then disabling
    # password authorization in favor of private key auth.
    ssh:
      # The name or the IP of the host
      hostname: xxx.xxx.xxx.xxx
      # The user to ssh as
      user: root
      # Password
      password: "pwd01"
      # This will add "sudo" to all commands, so 
      # you can login as non-root
      become: sudo
      
    # Labels assigned to the host. You can selects groups
    # of hosts using their labels
    labels:
      - db
  
    # Node specific properties, key=value
    properties:
      dbPassword: "pwd"
      
    # Node specific configuration options. YAML object. These override 
    # matching global configuration items
    configuration:


# Labels section assign labels to hosts
labels:
  # This assigns 'controller' label to 'host1' and 'host2'
  controller:
     - host1
     - host2
  # This assigns 'worker' label to 'host3' and 'host4'
  worker:
     - host3
     - host4
     
# Configuration section contains global configuration options
configuration:
  endpoint: http://myendpoint
  dbport: 2222
```


Watermelon merges all configuration files and the contents of the
`configuration` item in the inventory, and serves them as a common
configuration tree. You can query individual items using JSON
pointers, and unmarshal them into structs.

For example, the following
will unmarshal the contents of JSON location `/endpoint` :

```
var endpoint string
session.GetCfg("/endpoint", &endpoint)
// endpoint is now http://myendpoint
```


Watermelon client runtime provides APIs to access inventory items by
their labels or ids. For example:

```
// Select all hosts that has 'controller' label, and run 
// the function for each host. Each func runs in it own
// goroutine.
session.ForAllSelected(client.Has("controller"), func(host client.Host) error {
		var cfg MyConfig
        // Unmarshal contents of /configuration/myconfig from
        // the inventory file, or /myconfig
        // from one of the configuration files
		session.GetCfg("/myconfig", &cfg)
        ...
```
        
## Logs

For each run, Watermelon server creates a log directory containing the
timestamp of the run. Under that directory, each remote host and if
you use it `localhost` has separate log files. Any logs generated for
remote hosts and localhost will be written to its corresponding file.

## Modules

You write your infrastructure code as `modules`. A module is an
executable program that communicates with the watermelon server using
gRPC. Currently watermelon has a Go client runtime, but other runtimes
can be written for any language that supports gRPC.

A module source tree should look like the following:

```
moduleroot/
   moduleName/
      module.w
      ...
```

The `moduleroot` will be given to `watermelon` with the `--mdir` flag. Each
directory with a file called `module.w` is a module, and the module
name is the directory name it is under.

`module.w` is a shell script that is run in the module directory.
It can be executed in one of the following forms:

```
./module.w buildrun :port --log <loglevel>
```

`module.w` should build and run the module. When run, the module 
connects to the Watermelon server at `localhost:port`.

```
./module.w run :port --log <loglevel>
```

`module.w` should run the module, it doesn't have to build it.

For example, `pkg/module.w` file looks like the following. It runs
`make` if necessary, and then runs the binary with the arguments:

```
#!/bin/sh

if test "$#" -lt 2; then
    return 2
fi

if [ "$1" == "buildrun" ]; then
    make || exit 1;
fi

shift

./pkg $*
```



Watermelon server executes the module with a host:port
argument that the module uses to connect to the server. 

The following show how this is done in Go:
```
package main

import (
	"os"

	"github.com/bserdar/watermelon/client"
)

func main() {
	client.Run(os.Args[1:], nil, nil)
}
```


In order to make a function acessible from other modules, you have to
*export* it:

```
// Export the function SetupDB as smtp.SetupDB.
// Callers will have to call this function as module.smtp.SetupDB
var _ = client.Export("smtp.SetupDB", SetupDB)

// The SetupDB function, 
func SetupDB(session *client.Session) {
  ...
}
```

Exported functions can have one of the following signatures:

```
func WithInputAndOutput(*Session,InStruct) (OutStruct,error)
func WithInputOnly(*Session,InStruct) error
func WithInputNoError(*Session,InStruct)
func OnlySession(*Session)
```

The input and output structures must be JSON marshalable.

The `Session` provides the interface to the watermelon server. Using
the `Session`, the function can select hosts, and run commands on them.

```
session.ForAllSelected(client.Has("controller"), func(host client.Host) error {
        var ca []byte
        ca:=getCertificate(host)
        
        // Write the byte array to the remote host
        host.WriteFile("/certs", 0644, ca)
        // Run systemctl on the remote host
        host.Command("systemctl restart myservice")

        // Ensure directory exists on remote host
   	    host.Ensure("/etc/myapp/config", client.Ensure{}.EnsureDir())

        // Evaluate the Go template from local directory using templateData, and write it to remote host if different.
  	    changed, err := node.WriteFileFromTemplateFile("/etc/config", 0644, "templates/config", templateData)
  }
}
```

You can call an exported function from a module:

```
response:=session.Call("myModule","smtp.SetupDB",nil)
if !response.Success {
   return errors.New(response.ErrorMsg)
}
```

If the exported function gets an argument, you have to send an object
that can be JSON-marshaled:

```
session.Call("pkg","func",map[string]interface{}{"hostId":host.ID,"pkg":"ntpd"})
```

### Using gRPC to Export Functions

You can implement a gRPC server for the modules. The functions
exported using the gRPC server can be accessed using the
`session.Call` method as well as using direct gRPC.

To implement the module as a gRPC server, change the main as follows:

```
import (
	"os"

	"google.golang.org/grpc"

	"github.com/bserdar/watermelon-modules/pkg/yum"
	"github.com/bserdar/watermelon/client"
)

func main() {
    // Create the grpc server 
	yumServer := yum.Server{}
    // Register the grpc server to runtime 
	client.Run(os.Args[1:], nil, func(server *grpc.Server, rt *client.Runtime) {
        // Register the grpc server to the runtime with name "yum".
        // This allows other modules to call the gRPC functions of this module 
        // using session.Call("pkg","yum.funcName")
		rt.RegisterGRPCServer(&yumServer, "yum")
        
        // Register the gRPC server to the watermelon runtime gRPC server
		yum.RegisterYumServer(server, yumServer)
	})
}
```

This does two things: first it registers the gRPC server of the
implementation to the watermelon client runtime so it can proxy
non-gRPC calls to the gRPC server functions. Second, it registers a
gRPC server to the module's gRPC implementation.

The `main` function can register as many gRPC servers as you need.

The actual implementation of the module functions is as follows:

```
package yum

import (
	"github.com/bserdar/watermelon/client"
	"github.com/bserdar/watermelon/server/pb"
)

// This is the gRPC server for the yum module
type Server struct {
	client.GRPCServer
}

// The gRPC implementation of the Install function
func (s Server) Install(ctx context.Context, req *PackageParams) (*pb.Response, error) {
    // Get the watermelon session from the server
	session := s.SessionFromContext(ctx)
    // Do the work
    ...
    // Return the response
	return &pb.Response{Data: stdout, ErrorMsg: stderr, Modified: true}, nil
}
```

The session information is send using gRPC metadata, and
`s.SessionFromContext(ctx)` retrieves that.

If a module implements a gRPC server, then you can call the functions
either by `session.Call`, or by the gRPC bridge:

```
// Using session.Call
response := node.S.Call("pkg", "yum.Ensure", map[string]interface{}{"hostId": node.ID,
  "pkgs":    []string{"mongodb-org", "firewalld"},
  "version": "installed"})
```


When calling a module function using thr gRPC bridge, make sure you pass in a 
context obtained from `session.Context()`:
```
// Using gRPC bridge

// This will load the "pkg" module, and return a gRPC connection to it
pkgConn, err := session.ConnectModule("pkg")
if err != nil {
   return err
}
yumCli := yum.NewYumClient(pkgConn)
// You have to pass in the context you obtained from the session, otherwise the
// receiving end will not receive session info
yumCli.Ensure(node.S.Context(), &yum.EnsureParams{HostId: host.ID,
                Pkgs:    []string{"wget", "ntpdate", "iptables-services"},
                Version: "installed"})
```



## Examples

Install ETCD to all nodes if it is not already installed:

```
func ETCDBootstrap(session *client.Session) {
  // Select all hosts that are labeled with etcd
  session.ForAllSelected(client.Has("etcd"), func(host client.Host) error {
      // This section runs for all hosts with label "etcd" concurrently
      
     // Check if etcd already exists
     rsp := host.Commandf("/usr/local/bin/etcd -version")
     if strings.Index(string(rsp.Stdout), "etcd Version") == -1 {
        // Download etcd tarfile to /tmp and install
        rsp := host.Command("mktemp -d")
        tempDir := strings.TrimSpace(string(rsp.Stdout))
        rsp = host.Commandf("wget %s -O %s/tarfile", ETCDTarFile, tempDir)
        if rsp.ExitCode != 0 {
          return fmt.Errorf("Cannot wget etcd: %s", string(rsp.Stderr))
        }
        // Print log message for this host
        host.Logf("Downloaded etcd")
        host.Commandf("sh -c 'cd %s && tar  --strip-components=1 -x -f tarfile && mv etcd* /usr/local/bin'", tempDir)
     }
     return nil
  })
}
```

Reset mariadb root password:

```
func ResetRootPWD(session *client.Session) {
  // Run this for all hosts labeled with "mariadb"
  session.ForAllSelected(client.Has("mariadb"), func(host client.Host) error {
    // Get mariadb root password from host properties in inventory
    password, ok := host.GetInfo().Properties["mariadb_root_password"]
    if !ok {
      return fmt.Errorf("Root password required in mariadb_root_password")
    }
    host.Command("systemctl stop mariadb")
    host.Command(`nohup mysqld_safe --skip-grant-tables </dev/null >/dev/null 2>/dev/null &`)
    time.Sleep(2 * time.Second)
    host.Commandf(`cat <<EOF | mysql -u root
use mysql;
update user SET PASSWORD=PASSWORD("%s") WHERE USER='root';
flush privileges;
exit
EOF
`, password)
    host.Commandf("mysqladmin -u root --password=%s shutdown", rootPwd)
    host.Command("systemctl start mariadb")
  })
}
```

Write /etc/hosts from a template:

```
func WriteHosts(session *client.Session) {
  // For all hosts
  session.ForAllSelected(client.AllHosts,func(host client.Host) error {
    // Get host information for all hosts
    hosts := host.S.GetHosts(client.AllHosts)
    // Write /etc/hosts from the template. The templates are relative to the module root
    return node.WriteFileFromTemplateFile("/etc/hosts", 0644, "templates/hosts", hosts)
  })
}
```

where the hosts template is:
```
27.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4
::1         localhost localhost.localdomain localhost6 localhost6.localdomain6

{{range $index,$host := .}}
{{range .Addresses}}
{{.Address}} {{$host.ID}}
{{end -}}
{{end -}}
```
