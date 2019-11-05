package inventory

import (
	"context"

	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/pb"
)

type srv struct {
}

// Selects a subset of an inventory based on the given criteria, and
// returns a new inventory id representing the subset
func (s srv) Select(ctx context.Context, req *pb.InvSelectRequest) (*pb.InvId, error) {
	session := server.GetSession(req.Session)
	if session == nil {
		return nil, server.ErrInvalidSession(req.Session)
	}
	selectors := make([]*pb.Selector, len(req.Sel))
	for i, x := range req.Sel {
		selectors[i] = x
	}
	id, err := session.GetInv().Select(string(req.From), selectors...)
	if err != nil {
		return nil, err
	}
	return &pb.InvId{ID: string(id)}, nil
}

// Union takes a union of a set of inventories and returns a new inventory
// containing all the hosts in the combined inventories
func (s srv) Union(ctx context.Context, req *pb.InvUnionRequest) (*pb.InvId, error) {
	session := server.GetSession(req.Session)
	if session == nil {
		return nil, server.ErrInvalidSession(req.Session)
	}
	ids := make([]string, len(req.Sources))
	for i, x := range req.Sources {
		ids[i] = string(x)
	}
	id, err := session.GetInv().Union(ids)
	if err != nil {
		return nil, err
	}
	return &pb.InvId{ID: string(id)}, nil
}

// Make creates a new inventory containing the given hosts
func (s srv) Make(ctx context.Context, req *pb.HostIds) (*pb.InvId, error) {
	session := server.GetSession(req.Session)
	if session == nil {
		return nil, server.ErrInvalidSession(req.Session)
	}
	id, err := session.GetInv().Make(req.HostIds)
	if err != nil {
		return nil, err
	}
	return &pb.InvId{ID: string(id)}, nil
}

// Adds new hosts to an inventory
func (s srv) Add(ctx context.Context, req *pb.InvAddRequest) (*pb.InvId, error) {
	session := server.GetSession(req.Session)
	if session == nil {
		return nil, server.ErrInvalidSession(req.Session)
	}

	id, err := session.GetInv().Add(string(req.Inv), req.Hosts.IDs)
	if err != nil {
		return nil, err
	}
	return &pb.InvId{ID: string(id)}, nil
}

// Returns the host ids for all the hosts in an inventory
func (s srv) GetHostIds(ctx context.Context, req *pb.InvIdRequest) (*pb.HostIds, error) {
	session := server.GetSession(req.Session)
	if session == nil {
		return nil, server.ErrInvalidSession(req.Session)
	}
	ids, err := session.GetInv().GetHostIDs(string(req.ID))
	if err != nil {
		return nil, err
	}
	hids := make([]string, len(ids))
	for i, x := range ids {
		hids[i] = string(x)
	}
	return &pb.HostIds{HostIds: hids}, nil
}

// Returns information about some hosts
func (s srv) GetHostInfo(ctx context.Context, req *pb.HostIds) (*pb.HostInfos, error) {
	session := server.GetSession(req.Session)
	if session == nil {
		return nil, server.ErrInvalidSession(req.Session)
	}
	h, err := session.GetInv().GetHostInfo(req.HostIds)
	if err != nil {
		return nil, err
	}
	ret := pb.HostInfos{HostInfos: make([]*pb.HostInfo, len(h))}
	for i, x := range h {
		ret.HostInfos[i] = x
	}
	return &ret, nil
}

// Returns hosts in an inventory
func (s srv) GetHosts(ctx context.Context, req *pb.InvIdRequest) (*pb.HostInfos, error) {
	ids, err := s.GetHostIds(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.GetHostInfo(ctx, &pb.HostIds{HostIds: ids.HostIds, Session: req.Session})
}

// Release notifies the server that this inventory is no longer needed
func (s srv) Release(ctx context.Context, req *pb.InvIdRequest) (*pb.Empty, error) {
	session := server.GetSession(req.Session)
	if session == nil {
		return nil, server.ErrInvalidSession(req.Session)
	}

	session.GetInv().Release(string(req.ID))
	return &pb.Empty{}, nil
}

// NewServer returns a new inventory grpc server
func NewServer() pb.InventoryServer {
	return srv{}
}

func isMatch(in *pb.PropertySet_KVS, h *pb.HostInfo) bool {
	if v, ok := h.Properties[in.Key]; ok {
		for _, x := range in.Values {
			if x == v {
				return true
			}
		}
	}
	return false
}

func IsMatch(s *pb.Selector, h *pb.HostInfo) bool {
	switch k := s.Select.(type) {
	case *pb.Selector_HasAllLabels:
		for _, r := range k.HasAllLabels.Labels {
			if !server.HasLabel(h, r) {
				return false
			}
		}
		return true
	case *pb.Selector_HasAnyLabel:
		for _, r := range k.HasAnyLabel.Labels {
			if server.HasLabel(h, r) {
				return true
			}
		}
		return false
	case *pb.Selector_HasNoneLabels:
		for _, r := range k.HasNoneLabels.Labels {
			if server.HasLabel(h, r) {
				return false
			}
		}
		return true
	case *pb.Selector_ByID:
		for _, r := range k.ByID.IDs {
			if h.ID == r {
				return true
			}
		}
		return false
	case *pb.Selector_HasAnyProperty:
		for _, p := range k.HasAnyProperty.Properties {
			if isMatch(p, h) {
				return true
			}
		}
		return false
	case *pb.Selector_HasAllProperty:
		for _, p := range k.HasAllProperty.Properties {
			if !isMatch(p, h) {
				return false
			}
		}
		return true
	}
	return false
}
