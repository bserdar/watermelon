package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/inventory"
	"github.com/bserdar/watermelon/server/logging"
	"github.com/bserdar/watermelon/server/module"
	"github.com/bserdar/watermelon/server/pb"
	"github.com/bserdar/watermelon/server/remote"
)

var runArgs = struct {
	mdir   []string
	grpc   string
	logdir string
	stdout bool
	config string
}{}

func init() {
	runCmd.PersistentFlags().StringSliceVar(&runArgs.mdir, "mdir", []string{"."}, "Directory containing modules. You can specify this flag multiple times for each directory. This is combined with WM_MODULES env var.")
	runCmd.Flags().StringVar(&runArgs.grpc, "listen", "localhost:9876", "GRPC port to listen")
	runCmd.Flags().StringVar(&runArgs.logdir, "log", "./log", "Log directory")
	runCmd.Flags().BoolVar(&runArgs.stdout, "stdout", false, "Log to stdout as well")
	runCmd.Flags().StringVar(&runArgs.config, "cfg", "", "Configuration file.")
	rootCmd.AddCommand(runCmd)
}

// runCmd runs a module
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a function in a package",
	Long:  `Run a function in a package. Pass the package name, function name, and any additional args`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		moduleDirs := strings.Split(os.Getenv("WM_MODULES"), fmt.Sprint(os.PathListSeparator))

		logLevel := "info"
		if verbose {
			log.SetLevel(log.DebugLevel)
			logLevel = "debug"
		} else {
			log.SetLevel(log.InfoLevel)
		}
		log.Debug("Initializing...")

		var cfg interface{}
		var err error
		if len(runArgs.config) > 0 {
			cfg, err = readConfig(runArgs.config)
			if err != nil {
				panic(err)
			}
		}

		session := server.NewSession()
		session.SetLogStdout(runArgs.stdout)
		defer session.Close()
		session.SetConfig(cfg)

		inv, config := loadInventory(session)
		session.SetInv(inv)
		if config != nil {
			session.SetConfig(mergeConfig(session.GetConfig(), config))
		}
		logdir := logging.GetLogDir(runArgs.logdir, args[0])
		session.SetLog(logging.Logging{Logdir: logdir})
		fmt.Printf("Logs are under %s\n", logdir)
		os.MkdirAll(logdir, 0775)

		grpcServer := grpc.NewServer()
		grpcListener, err := net.Listen("tcp", runArgs.grpc)
		if err != nil {
			panic(err)
		}
		defer grpcListener.Close()

		// Setup the session
		lmgr := module.NewLifecycleManager()
		lmgr.RunModuleScript = func(first bool, dir string) error {
			t := "run"
			if first {
				t = "buildrun"
			}
			log.Debugf("Calling %s first: %v", dir, first)
			return lmgr.ExecModule("/bin/sh", "-c", fmt.Sprintf("cd %s;/bin/sh ./module.w %s %s --log %s", dir, t, runArgs.grpc, logLevel))
		}
		lmgr.LocalModuleFunc = server.CallLocalModule
		session.SetModules(lmgr)
		lmgr.ModuleLookupDirs = append(runArgs.mdir, moduleDirs...)
		log.Debugf("Module search dirs: %v", lmgr.ModuleLookupDirs)
		pb.RegisterLifecycleServer(grpcServer, lmgr)
		pb.RegisterInventoryServer(grpcServer, inventory.NewServer())
		pb.RegisterRemoteServer(grpcServer, remote.New())

		go func() {
			log.Debugf("Listening at %s", runArgs.grpc)
			grpcServer.Serve(grpcListener)
			log.Debugf("Closed server at %s", runArgs.grpc)
		}()
		// Give the server a chance to start
		time.Sleep(time.Millisecond * 100)

		// Run the module, with optional args
		session.SetArgs(args[2:])
		log.Debugf("Calling %s.%s", args[0], args[1])
		_, err = session.GetModules().SendRequest(session.GetID(), args[0], args[1], nil)
		log.Debugf("Result of main: %v", err)
		if err != nil {
			log.Errorf(err.Error())
		}
		log.Debugf("Closing session")
		session.Close()
	}}

func readConfig(config string) (interface{}, error) {
	data, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, err
	}
	var v interface{}
	ext := strings.ToUpper(filepath.Ext(config))
	if ext == ".YML" || ext == ".YAML" {
		err := yaml.Unmarshal(data, &v)
		if err != nil {
			return nil, err
		}
		return server.MapYaml(v), nil
	}
	if ext == ".JSON" {
		err := json.Unmarshal(data, &v)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, fmt.Errorf("Unrecognized extension: %s", ext)
}

// Merge src into dest
func mergeConfig(dest, src interface{}) interface{} {
	if dest == nil {
		return src
	}
	if src == nil {
		return dest
	}
	if md, ok := dest.(map[string]interface{}); ok {
		if ms, ok := src.(map[string]interface{}); ok {
			for k, v := range ms {
				if ov, ok := md[k]; ok {
					md[k] = mergeConfig(ov, v)
				} else {
					md[k] = v
				}
			}
		} else {
			panic(fmt.Sprintf("Incompatible configs: %v - %v", dest, src))
		}
		return server.MapYaml(md)
	}

	if as, ok := src.([]interface{}); ok {
		return server.MapYaml(as)
	}
	return server.MapYaml(src)
}
