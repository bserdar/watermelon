package server

import (
	"github.com/google/uuid"

	"github.com/bserdar/watermelon/server/pb"
	sshdial "github.com/bserdar/watermelon/server/ssh"
)

// AllHosts is the inventory that contains all known hosts
var AllHosts = "all"

// Inventory server interface. Inventory keeps all known servers
type Inventory interface {
	// Select returns a new inventory containing only the hosts
	// selected by the given selectors from the given inventory
	Select(string, ...*pb.Selector) (string, error)

	// Union combines inventories to build a new host set containing
	// all the hosts in the sources
	Union([]string) (string, error)

	// Make creates a new inventory from the given host IDs
	Make([]string) (string, error)

	// Add adds new hosts to the given inventory, and returns the new
	// inventory ID
	Add(string, []string) (string, error)

	// GetHostIDs returns the host IDs included in the inventory
	GetHostIDs(string) ([]string, error)

	// GetHostInfo returns the host information for the given hosts
	GetHostInfo([]string) ([]*pb.HostInfo, error)

	// Release notifies the server that the inventory is no longer
	// needed, and can be freed
	Release(string)
}

type InternalInventory interface {
	Inventory

	GetHost([]string) ([]*Host, error)
}

// NewInventoryID returns a new inventory ID. It skips AllHosts.
func NewInventoryID() string {
	return uuid.New().String()
}

// InventoryConfiguration contains the configuration loaded from the inventory
type InventoryConfiguration struct {
	PrivateKey *sshdial.RawPrivateKey
}
