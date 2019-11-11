package yml

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/bserdar/watermelon/server"
	"github.com/bserdar/watermelon/server/pb"
	sshdial "github.com/bserdar/watermelon/server/ssh"
)

// Inventory contains the hosts and labels as parsed from YML
type Inventory struct {
	PrivateKeyFile string              `yaml:"privateKey,omitempty"`
	Passphrase     string              `yaml:"passphrase,omitempty"`
	Configuration  interface{}         `yaml:"configuration,omitempty"`
	Hosts          []Host              `yaml:"hosts,omitempty"`
	Labels         map[string][]string `yaml:"labels,omitempty"`
}

// SSH specifics
type SSH struct {
	Hostname string `yaml:"hostname"`
	Network  string `yaml:"network"`
	Port     int    `yaml:"port,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	Become   string `yaml:"become,omitempty"`
}

// Host defines a YAML host
type Host struct {
	Network    string            `yaml:"network,omitempty"`
	ID         string            `yaml:"id,omitempty"`
	Labels     []string          `yaml:"labels,omitempty"`
	Properties map[string]string `yaml:"properties,omitempty"`
	Address    string            `yaml:"address,omitempty"`
	Addresses  []struct {
		Address string `yaml:"address"`
		Name    string `yaml:"name"`
	} `yaml:"addresses"`
	SSH           *SSH        `yaml:"ssh,omitempty"`
	Configuration interface{} `yaml:"configuration,omitempty"`
}

func (h *Host) toHost() (*server.Host, error) {
	host := &server.Host{HostInfo: pb.HostInfo{ID: h.ID,
		Labels:     h.Labels,
		Properties: h.Properties}}
	if h.SSH != nil {
		host.Hostname = h.SSH.Hostname
		if len(h.SSH.Hostname) == 0 {
			return nil, errors.New("Empty hostname")
		}
		host.Network = h.SSH.Network
		host.Port = h.SSH.Port
		host.LoginUser = h.SSH.User
		host.LoginPassword = h.SSH.Password
		host.Become = h.SSH.Become
	}
	host.Configuration = server.MapYaml(h.Configuration)

	host.Defaults()
	if len(h.Address) == 0 {
		if len(h.Addresses) == 0 {
			// Need to discover addresses
			err := host.DiscoverIPs()
			if err != nil {
				return nil, err
			}
		} else {
			for _, a := range h.Addresses {
				ip := net.ParseIP(a.Address)
				if ip == nil {
					return nil, fmt.Errorf("Cannot parse address %s", a.Address)
				}
				if len(a.Name) == 0 {
					return nil, fmt.Errorf("Name required for address %s", a.Address)
				}
				host.Addresses = append(host.Addresses, &pb.Address{Name: a.Name, Address: ip.String()})
			}
		}
	} else {
		if len(h.Addresses) == 0 {
			ip := net.ParseIP(h.Address)
			if ip == nil {
				return nil, fmt.Errorf("Cannot parse address %s", h.Address)
			}
			host.Addresses = append(host.Addresses, &pb.Address{Name: server.Primary, Address: ip.String()})
		} else {
			return nil, fmt.Errorf("Both address and addresses are given for %s", h.ID)
		}
	}

	return host, nil
}

// ToInventory converts the YAML inventory to a host array
func (i *Inventory) ToInventory() ([]*server.Host, error) {
	ret := make([]*server.Host, 0, len(i.Hosts))
	for _, h := range i.Hosts {
		host, err := h.toHost()
		if err != nil {
			return nil, err
		}
		ret = append(ret, host)
	}
	errors := make([]string, 0)
	find := func(id string) *server.Host {
		for _, h := range ret {
			if h.ID == id {
				return h
			}
		}
		return nil
	}
	for k, l := range i.Labels {
		for _, h := range l {
			host := find(h)
			if host == nil {
				errors = append(errors, h)
			}
			host.Labels = append(host.Labels, k)
		}
	}
	if len(errors) > 0 {
		return nil, fmt.Errorf("These hosts are referenced in the inventory, but they are not defined: %s", strings.Join(errors, ","))
	}

	for _, x := range ret {
		x.Backend = server.GetBackend("linux", x)
	}
	return ret, nil
}

// LoadInventory loads inventory from a yaml file
func LoadInventory(f string) (server.InventoryConfiguration, []*server.Host, map[string]interface{}, error) {
	var inv Inventory
	var cfg server.InventoryConfiguration
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return cfg, nil, nil, err
	}

	err = yaml.Unmarshal(data, &inv)
	if err != nil {
		return cfg, nil, nil, err
	}
	if len(inv.PrivateKeyFile) > 0 {
		pk, err := ioutil.ReadFile(inv.PrivateKeyFile)
		if err != nil {
			return cfg, nil, nil, err
		}
		cfg.PrivateKey = &sshdial.RawPrivateKey{PEMData: pk, Passphrase: inv.Passphrase}
	}
	hi, err := inv.ToInventory()
	if cfg.PrivateKey != nil {
		for _, host := range hi {
			host.KeyAuth = cfg.PrivateKey
		}
	}

	configuration := server.MapYaml(inv.Configuration)
	if configuration == nil {
		configuration = make(map[string]interface{})
	}
	return cfg, hi, configuration.(map[string]interface{}), err
}
