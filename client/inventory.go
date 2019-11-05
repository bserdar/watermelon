package client

import (
	"context"

	"github.com/bserdar/watermelon/server/pb"
)

// AllHosts of the inventory
var AllHosts = "all"

// Inventory is the inventory runtime implementation for clients
type Inventory struct {
	impl pb.InventoryClient
}

// Select a subset of an inventory based on the given criteria, and
// returns a new inventory id representing the subset
func (inv Inventory) Select(session string, from string, what ...Selector) (string, error) {
	selectors := make([]*pb.Selector, len(what))
	for i, x := range what {
		selectors[i] = selectorToPb(x)
	}
	invid, err := inv.impl.Select(context.Background(), &pb.InvSelectRequest{Session: session,
		From: string(from),
		Sel:  selectors})
	if err != nil {
		return string(""), err
	}
	return string(invid.ID), nil
}

// Union combines inventories to build a new host set containing
// all the hosts in the sources
func (inv Inventory) Union(session string, what []string) (string, error) {
	invids := make([]string, len(what))
	for i, x := range what {
		invids[i] = string(x)
	}
	id, err := inv.impl.Union(context.Background(), &pb.InvUnionRequest{Session: session,
		Sources: invids})
	if err != nil {
		return string(""), err
	}
	return string(id.ID), nil
}

// Make creates a new inventory from the given host IDs
func (inv Inventory) Make(session string, from []string) (string, error) {
	id, err := inv.impl.Make(context.Background(), &pb.HostIds{Session: session,
		HostIds: from})
	if err != nil {
		return string(""), err
	}
	return string(id.ID), nil
}

// Add adds new hosts to the given inventory, and returns the new
// inventory ID
func (inv Inventory) Add(session string, to string, hosts []string) (string, error) {
	id, err := inv.impl.Add(context.Background(), &pb.InvAddRequest{Session: session,
		Inv:   string(to),
		Hosts: &pb.HostIdSet{IDs: hosts}})
	if err != nil {
		return string(""), err
	}
	return string(id.ID), nil
}

// GetHostIDs returns the host IDs included in the inventory
func (inv Inventory) GetHostIDs(session string, invID string) ([]string, error) {
	ids, err := inv.impl.GetHostIds(context.Background(), &pb.InvIdRequest{Session: session, ID: string(invID)})
	if err != nil {
		return nil, err
	}
	return ids.HostIds, nil
}

// GetHostInfo returns the host information for the given hosts
func (inv Inventory) GetHostInfo(session string, IDs []string) ([]*pb.HostInfo, error) {
	hi, err := inv.impl.GetHostInfo(context.Background(), &pb.HostIds{HostIds: IDs,
		Session: session})
	if err != nil {
		return nil, err
	}
	ret := make([]*pb.HostInfo, len(hi.HostInfos))
	for i, x := range hi.HostInfos {
		if x.Properties == nil {
			x.Properties = make(map[string]string)
		}
		ret[i] = x
	}
	return ret, nil
}

// GetHosts returns hosts in an inventory
func (inv Inventory) GetHosts(session string, invID string) ([]*pb.HostInfo, error) {
	hi, err := inv.impl.GetHosts(context.Background(), &pb.InvIdRequest{Session: session, ID: string(invID)})
	if err != nil {
		return nil, err
	}
	ret := make([]*pb.HostInfo, len(hi.HostInfos))
	for i, x := range hi.HostInfos {
		if x.Properties == nil {
			x.Properties = make(map[string]string)
		}
		ret[i] = x
	}
	return ret, nil
}

// Release notifies the server that the inventory is no longer
// needed, and can be freed
func (inv Inventory) Release(session string, id string) {
	inv.impl.Release(context.Background(), &pb.InvIdRequest{Session: session, ID: string(id)})
}

func makeKeyValues(in []KeyAndValues) *pb.PropertySet {
	p := make([]*pb.PropertySet_KVS, len(in))
	for i, x := range in {
		p[i] = &pb.PropertySet_KVS{Key: x.Key, Values: x.Values}
	}
	return &pb.PropertySet{Properties: p}
}

func selectorToPb(sel Selector) *pb.Selector {
	switch k := sel.(type) {
	case HasAllLabels:
		return &pb.Selector{Select: &pb.Selector_HasAllLabels{HasAllLabels: &pb.LabelSet{Labels: k.Labels}}}
	case HasAnyLabel:
		return &pb.Selector{Select: &pb.Selector_HasAnyLabel{HasAnyLabel: &pb.LabelSet{Labels: k.Labels}}}
	case HasNoneLabels:
		return &pb.Selector{Select: &pb.Selector_HasNoneLabels{HasNoneLabels: &pb.LabelSet{Labels: k.Labels}}}
	case SelectByID:
		return &pb.Selector{Select: &pb.Selector_ByID{ByID: &pb.HostIdSet{IDs: k.IDs}}}
	case SelectByAnyProperty:
		return &pb.Selector{Select: &pb.Selector_HasAnyProperty{HasAnyProperty: makeKeyValues(k.Any)}}
	case SelectByAllProperty:
		return &pb.Selector{Select: &pb.Selector_HasAllProperty{HasAllProperty: makeKeyValues(k.All)}}
	}
	return nil
}
