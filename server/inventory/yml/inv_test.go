package yml

import (
	"testing"

	yml "gopkg.in/yaml.v2"

	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/inventory"
	"github.com/bserdar/watermelon/server/pb"
)

func TestInventory(t *testing.T) {
	in := `---
hosts:
  - id: h1
    address: 127.0.0.2
    labels:
      - l1
      - l2
  - id: h2
    address: 127.0.0.3
  - id: h3
    address: 127.0.0.4
    labels:
      - lx
labels:
  bulk1:
    - h2
    - h3
`
	var invd Inventory
	err := yml.Unmarshal([]byte(in), &invd)
	if err != nil {
		t.Errorf("Cannot parse: %s", err)
		return
	}
	i, err := invd.ToInventory()
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if len(i) != 3 {
		t.Errorf("Expecting 3 hosts, got %d", len(i))
	}

	srv := inventory.NewInvServer(i, server.InventoryConfiguration{}, nil)

	h1, _ := srv.Select(server.AllHosts, &pb.Selector{Select: &pb.Selector_ByID{ByID: &pb.HostIdSet{IDs: []string{"h1"}}}})
	al, _ := srv.GetHostIDs(h1)
	if len(al) == 0 {
		t.Errorf("Cannot find h1")
	}

	host, _ := srv.GetHostInfo(al)
	if server.ArrayContains(host[0].Labels, "bulk1") {
		t.Errorf("Not expecting label")
	}
	h2, _ := srv.Select(server.AllHosts, &pb.Selector{Select: &pb.Selector_ByID{ByID: &pb.HostIdSet{IDs: []string{"h2"}}}})
	al, _ = srv.GetHostIDs(h2)
	host, _ = srv.GetHostInfo(al)
	if !server.ArrayContains(host[0].Labels, "bulk1") {
		t.Errorf("Expecting label")
	}
	h3, _ := srv.Select(server.AllHosts, &pb.Selector{Select: &pb.Selector_ByID{ByID: &pb.HostIdSet{IDs: []string{"h3"}}}})
	al, _ = srv.GetHostIDs(h3)
	host, _ = srv.GetHostInfo(al)
	if !server.ArrayContains(host[0].Labels, "bulk1") {
		t.Errorf("Expecting label")
	}
}
