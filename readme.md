# Watermelon &nbsp; [![GoDoc](https://godoc.org/github.com/bserdar/watermelon?status.svg)](https://godoc.org/github.com/bserdar/watermelon)


Watermelon is an imperative configuration management tool. It lets you
write configuration scripts using a real programming language to
manage remote machines. Since you can use any programming language,
writing scripts with watermelon is similar to writing code to run
commands, however watermelon hides the details of connecting to those
machines, running remote commands, moving/copying files remotely. If a
watermalon script is run for multiple machines, that script can
execute for all machines simultaneously.


Watermelon is written in Go, and currently supports configuration
scripts written in Go. You can write scripts in any language that
support GRPC by implementing a runtime client for that language. Once
a runtime is implemented, a script written in one language can call
the functions written in another language.

## Install

Use `go get`:

```
go get github.com/bserdar/watermelon
```

Get modules:

```
git clone github.com/bserdar/watermelon-modules
```

You will need to pass the directory of these modules to `wm` to use
them, or define WM_MODULES environment variable:

```
export WM_MODULES="~/go/src/github.com/bserdar/watermelon-modules"
```

## Running

Use the `wm` executable to run scripts. This is the Watermelon
server. The server manages connections to remote machines,
configurations, logging, and provides common functions you can call
from modules, such as copying files, or executing commands on remote
nodes. Run the server using:

```
wm run --cfg myconfig.yml --inv inventory.yml --mdir /dir-to-modules/watermelon-modules --mdir /other/module/dir someModule.someFunc
```

This will load `myconfig.yml` as the configuration file, and
`inventory.yml` as the inventory file. The configuration file is a
yaml file whose contents can be queried from the running scripts. For
example, the following will unmarshal the contents of `/myconfig` to `cfg`:

```
session.GetCfg("/myconfig", &cfg)
```

The inventory defines the remote hosts and how to access
them. Watermelon hides the details of accessing hosts from the
scripts. You can run a command, make sure a directory exists, and
write a remote using a template stored locally as follows:

```
host.Command("systemctl restart myservice")
host.Ensure("/etc/myapp/config", client.Ensure{}.EnsureDir())
changed, err := node.WriteFileFromTemplateFile("/etc/config", 0644, "templates/config", templateData)
```


The `mdir` argument defines module base directories. Watermelon will
search these directories to locate modules at runtime. Each module is
a directory under one of these module directories, and the module name
is the directory name. Each module exports functions that can be
invoked from the command line, or from other modules. The above
command will run `someFunc` in a module called `someModule`.

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

You write your configuration scripts as `modules`. The name of the
module is the last component of the directory it is in. To run a
module named `myModule`, Watermelon looks at the file
`<dir>/myModule/module.w` where `<dir>` is one of the directories
given with the `mdir` command line argument. `module.w` is a script
that is run with one of `buildrun` or `run` options. It is run in the
module directory.

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


## Writing modules

A module is an executable program that communicates with the
Watermelon server using GRPC. Currently Watermelon has a Go client
runtime, but other runtimes can be written for any language that
supports GRPC. Watermelon server executes the module with a host:port
argument that the module uses to connect to the server. After that the
module receives requests from the Watermelon server, executes them,
and sends the results back. The module itself can call other
modules. Watermelon will locate the called module, build it, run it,
and pass the call to that module. For this to work, when you write a
module you have to declare your exported functions to the runtime.


```
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
```

The above program declares three functions, and assigns names
to Go functions. Lets say this module is called `mymodule`, when you
run Watermelon like this:

```
wm run --mdir /mymoduledir mymodule db.Bootstrap
```

Watermelon will locate this module under /mymoduledir/mymodule, build
it, run it, and then call the `dbBootstrap` function.


The `dbBootstrap` function gets a pointer to a client session:

```
func dbBootstrap(session *client.Session) {
}
```

Using the client session, dbBootstrap can execute commands in remote
hosts (or localhost), copy files to/from remote hosts, and do these things
in parallel for all hosts:

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
