package server

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Funcdef describes an implementation of the function
type Funcdef struct {
	// ParamsFactory returns an object instance that will be unmarshaled
	// from json. It should me a pointer.
	ParamsFactory func() interface{}
	// This is the actual implementation function. The input interface
	// is what's returned from ParamsFactory
	Impl func(Session, interface{}) (Response, error)
}

// Dispatcher is the default module function dispatcher
type Dispatcher struct {
	Name  string
	funcs map[string]Funcdef
}

// NewDispatcher returns a new dispatcher with the given module name
func NewDispatcher(name string) *Dispatcher {
	return &Dispatcher{Name: name, funcs: make(map[string]Funcdef)}
}

// Add adds a new function to dispatcher
func (d *Dispatcher) Add(name string, f Funcdef) *Dispatcher {
	if d.funcs == nil {
		d.funcs = make(map[string]Funcdef)
	}
	d.funcs[name] = f
	return d
}

// Func dispatches the function call. This allows embedding Dispatcher into module
func (d Dispatcher) Func(session Session, name string, in []byte) (Response, error) {
	log.Debugf("Local dispatcher for %s", name)
	f, ok := d.funcs[name]
	if !ok {
		log.Debugf("Not found: %s", name)
		return Response{}, fmt.Errorf("%s: not found: %s", d.Name, name)
	}
	params := f.ParamsFactory()
	err := json.Unmarshal(in, params)
	if err != nil {
		log.Debugf("Cannot unmarshal params: %v", err)
		return Response{}, err
	}
	log.Debugf("Calling")
	r, err := f.Impl(session, params)
	if err != nil {
		return Response{}, err
	}
	r.FuncName = name
	return r, nil
}
