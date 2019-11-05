package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Module interface is to be implemented by all modules
type Module interface {
	// Func calls a module function
	Func(Session, string, []byte) (Response, error)
	// Short help about the module
	Describe() string
	// Long help about the module
	Help() string
}

// LocalModules keeps a map of all local modules
var LocalModules = map[string]Module{}

// RegisterModule registers a local module
func RegisterModule(name string, m Module) {
	LocalModules[name] = m
}

// CallLocalModule calls a local module if there is one registered
func CallLocalModule(session, module, funcname string, data []byte) (Response, bool, error) {
	log.Debugf("Call local module: %s.%s", module, funcname)
	if mod, ok := LocalModules[module]; ok {
		s := GetSession(session)
		if s == nil {
			log.Debugf("Cannot find session")
			return Response{}, true, fmt.Errorf("Invalid session %s while calling %s.%s", session, module, funcname)
		}
		log.Debugf("Calling function")
		rsp, err := mod.Func(s, funcname, data)
		return rsp, true, err
	}
	log.Debugf("Func not found")
	return Response{}, false, nil
}
