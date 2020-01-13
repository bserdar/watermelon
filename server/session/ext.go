package session

import (
	"github.com/bserdar/watermelon/server"
)

// Registered extensions
var (
	Extensions = map[string]func(server.Session) Extension{}
)

// Extension interface
type Extension interface {
}

// ConfigExpander is an extension that can expand config 'ref' objects
type ConfigExpander interface {
	ExpandConfig(in interface{}) interface{}
}
