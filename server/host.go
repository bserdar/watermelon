package server

import (
	"fmt"
	"net"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/bserdar/watermelon/server/pb"
	sshdial "github.com/bserdar/watermelon/server/ssh"
)

// Primary interface name.
const Primary = "primary"

// Primary4 IP4 interface name
const Primary4 = "primary4"

// Primary6 IP6 interface name
const Primary6 = "primary6"

// LocalhostID is the "localhost"
var LocalhostID = "localhost"

// Host contains information about a host in the inventory
type Host struct {
	sync.Mutex
	pb.HostInfo

	Configuration interface{}

	// The bastion host to use to connect this host
	Bastion *Host
	// SSH hostname
	Hostname string
	Network  string
	Port     int

	// If nonempty, this user is used to login
	LoginUser string
	// Password of the user
	LoginPassword string
	// If non-nil, the public key signer derived from the private key will be used
	KeyAuth *sshdial.RawPrivateKey
	// HostPublicKey is initialized if empty, validated if nonempty
	HostPublicKey ssh.PublicKey

	// Become user methos
	Become string

	Backend HostBackend
}

// GetNetwork returns the network to connect to this host
func (h *Host) GetNetwork() string {
	if h.Network == "" {
		return "tcp"
	}
	return h.Network
}

// GetHostAndPort returns the host and port to connect
func (h *Host) GetHostAndPort() string {
	if h.Port != 0 {
		return fmt.Sprintf("%s:%d", h.Hostname, h.Port)
	}
	return fmt.Sprintf("%s:22", h.Hostname)
}

// GetUserName returns the login user
func (h *Host) GetUserName() string {
	return h.LoginUser
}

// GetAuth returns the supported SSH auth methods
func (h *Host) GetAuth() ([]ssh.AuthMethod, error) {
	ret := make([]ssh.AuthMethod, 0)
	if h.KeyAuth != nil {
		s, err := h.KeyAuth.GetSigner()
		if err != nil {
			return nil, err
		}
		ret = append(ret, ssh.PublicKeys(s))
	}
	if len(h.LoginPassword) > 0 {
		ret = append(ret, ssh.Password(h.LoginPassword))
	}
	return ret, nil
}

// Via returns a bastion host
func (h *Host) Via() sshdial.Host {
	if h.Bastion != nil {
		return h.Bastion
	}
	return nil
}

type hostCtx struct {
	session HostSession
	ref     int
	host    *Host
}

// Ctx is host session context
type Ctx interface {
	New(Session) (HostSession, error)
	Close()
}

// NewCtx returns a new host context
func (h *Host) NewCtx() Ctx {
	return &hostCtx{host: h}
}

// New returns a new session, The returned session must be closed
func (ctx *hostCtx) New(s Session) (HostSession, error) {
	if ctx.session != nil {
		ctx.ref++
		return ctx.session, nil
	}
	var err error
	ctx.session, err = ctx.host.Backend.NewSession(s, ctx.host)
	if err != nil {
		return nil, err
	}
	ctx.ref++
	return ctx.session, nil
}

// Close a host context
func (ctx *hostCtx) Close() {
	if ctx.ref > 0 {
		ctx.ref--
	}
	if ctx.ref == 0 {
		ctx.session.Close()
		ctx.session = nil
	}
}

// Localhost is the localhost
var Localhost = &Host{HostInfo: pb.HostInfo{ID: LocalhostID,
	Labels:     make([]string, 0),
	Properties: make(map[string]string)}}

func init() {
	Localhost.DiscoverIPs()
}

// Defaults initializes uninitialized host fields with defaults
func (h *Host) Defaults() {
	if len(h.Network) == 0 {
		h.Network = "tcp"
	}
	if h.Port == 0 {
		h.Port = 22
	}
	if len(h.ID) == 0 {
		h.ID = h.Hostname
	}
	if h.Labels == nil {
		h.Labels = make([]string, 0)
	}
	if h.Properties == nil {
		h.Properties = make(map[string]string)
	}
}

// HasLabel returns true if host has the label l
func HasLabel(h *pb.HostInfo, l string) bool {
	return ArrayContains(h.Labels, l)
}

// DiscoverIPs initializes the host IPs
func (h *Host) DiscoverIPs() error {
	ips, error := net.LookupIP(h.Hostname)
	if error != nil {
		return error
	}

	n4 := 0
	n6 := 0
	var ip4, ip6 net.IP
	for _, ip := range ips {
		if ip.To4() == nil {
			if ip.To16() == nil {
				return fmt.Errorf("Invalid address: %s", ip.String())
			}
			n6++
			if n6 > 1 {
				return fmt.Errorf("Host has more than one IP6 address, you have to name them")
			}
			h.Addresses = append(h.Addresses, &pb.Address{Name: Primary6, Address: ip.String()})
			ip6 = ip
		} else {
			n4++
			if n4 > 1 {
				return fmt.Errorf("Host has more than one IP4 address, you have to name them")
			}
			h.Addresses = append(h.Addresses, &pb.Address{Name: Primary4, Address: ip.String()})
			ip4 = ip
		}
	}
	if n4 == 1 {
		h.Addresses = append(h.Addresses, &pb.Address{Name: Primary, Address: ip4.String()})
	}
	if n6 == 1 {
		h.Addresses = append(h.Addresses, &pb.Address{Name: Primary, Address: ip6.String()})
	}
	return nil
}

// WriteFile writes a file on the host
func (h *Host) WriteFile(ctx Ctx, s Session, name string, perms os.FileMode, content []byte) (CmdErr, error) {
	session, err := ctx.New(s)
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	return session.WriteFile(name, perms, content)
}

// ReadFile reads a file from the host
func (h *Host) ReadFile(ctx Ctx, s Session, name string) (os.FileInfo, []byte, CmdErr, error) {
	session, err := ctx.New(s)
	if err != nil {
		return nil, nil, nil, err
	}
	defer ctx.Close()
	return session.ReadFile(name)
}

// RunCmd runs cmd
func (h *Host) RunCmd(ctx Ctx, s Session, cmd string, env map[string]string) (HostCommandResponse, error) {
	session, err := ctx.New(s)
	if err != nil {
		return HostCommandResponse{}, err
	}
	defer ctx.Close()
	return session.Run(cmd, env)
}

// FileOwner contains owner information
type FileOwner struct {
	OwnerName string
	OwnerID   string
	GroupName string
	GroupID   string
}

// GetFileInfo retrieves file info from a host. If the file is
// missing, will return FileOwner{},nil,nil
func (h *Host) GetFileInfo(ctx Ctx, s Session, file string) (FileOwner, os.FileInfo, CmdErr, error) {
	log.Debugf("GetFileInfo enter file:%s ctx: %+v", file, ctx)
	session, err := ctx.New(s)
	if err != nil {
		return FileOwner{}, nil, nil, err
	}
	defer ctx.Close()
	return session.GetFileInfo(file)
}

// MkDir creates dir
func (h *Host) MkDir(ctx Ctx, s Session, path string) (CmdErr, error) {
	session, err := ctx.New(s)
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	return session.MkDir(path)
}

// Chmod changes mode
func (h *Host) Chmod(ctx Ctx, s Session, path string, mode int) (CmdErr, error) {
	session, err := ctx.New(s)
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	return session.Chmod(path, mode)
}

// Chown changes user/group, whichever is nonempty
func (h *Host) Chown(ctx Ctx, s Session, path string, user, group string) (CmdErr, error) {
	session, err := ctx.New(s)
	if err != nil {
		return nil, err
	}
	defer ctx.Close()
	return session.Chown(path, user, group)
}

// FileDesc describes the attributes of a file/directory
type FileDesc struct {
	Mode  *int
	UID   *string
	GID   *string
	User  *string
	Group *string
	Dir   *bool
}

// Ensure a file has the desired attributes
func (h *Host) Ensure(ctx Ctx, s Session, path string, desc FileDesc) (bool, CmdErr, error) {
	_, err := ctx.New(s)
	if err != nil {
		return false, nil, err
	}
	defer ctx.Close()

	log.Debugf("Ensure %s %+v ctx: %+v", path, desc, ctx)
	owner, fi, _, err := h.GetFileInfo(ctx, s, path)
	if err != nil {
		return false, nil, err
	}
	log.Debugf("Get file info result: %+v ctx:%+v", owner, ctx)
	changed := false
	if desc.Dir != nil {
		// Create a directory if it is not there
		if fi != nil && fi.IsDir() {
			log.Debugf("Already there")
		} else {
			cerr, err := h.MkDir(ctx, s, path)
			if err != nil {
				return false, nil, err
			}
			if cerr != nil {
				return false, cerr, nil
			}
			changed = true
			owner, fi, _, _ = h.GetFileInfo(ctx, s, path)
		}
	}

	if desc.Mode != nil {
		if fi != nil {
			perm := fi.Mode().Perm()
			if *desc.Mode == int(perm) {
				log.Debugf("Correct file mode")
			} else {
				cerr, err := h.Chmod(ctx, s, path, *desc.Mode)
				if err != nil {
					return false, nil, err
				}
				if cerr != nil {
					return false, cerr, nil
				}
				changed = true
				owner, fi, _, _ = h.GetFileInfo(ctx, s, path)
			}
		}
	}

	if desc.UID != nil || desc.User != nil {
		if (desc.UID != nil && *desc.UID == owner.OwnerID) ||
			(desc.User != nil && *desc.User == owner.OwnerName) {
			log.Debugf("Correct user")
		} else {
			var cerr CmdErr
			var err error
			if desc.UID != nil {
				cerr, err = h.Chown(ctx, s, path, *desc.UID, "")
			} else if desc.User != nil {
				cerr, err = h.Chown(ctx, s, path, *desc.User, "")
			}
			if err != nil {
				return false, nil, err
			}
			if cerr != nil {
				return false, cerr, nil
			}
			changed = true
			owner, fi, _, _ = h.GetFileInfo(ctx, s, path)
		}
	}
	if desc.GID != nil || desc.Group != nil {
		if (desc.GID != nil && *desc.GID == owner.GroupID) ||
			(desc.Group != nil && *desc.Group == owner.GroupName) {
			log.Debugf("Correct group")
		} else {
			var cerr CmdErr
			var err error
			if desc.GID != nil {
				cerr, err = h.Chown(ctx, s, path, "", *desc.GID)
			} else if desc.Group != nil {
				cerr, err = h.Chown(ctx, s, path, "", *desc.Group)
			}
			if err != nil {
				return false, nil, err
			}
			if cerr != nil {
				return false, cerr, nil
			}
			changed = true
			owner, fi, _, _ = h.GetFileInfo(ctx, s, path)
		}
	}

	return changed, nil, nil
}
