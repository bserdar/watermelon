package ssh

import (
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Host interface used by the ssh dialer.
type Host interface {
	GetID() string
	// Returns the network, "tcp"
	GetNetwork() string
	// Returns name:port
	GetHostAndPort() string
	// Returns ssh login username
	GetUserName() string

	// Return auth methods for the hots
	GetAuth() ([]ssh.AuthMethod, error)

	// Reach host via another one
	Via() Host
}

// Client wraps SSH client
type Client struct {
	SSH *ssh.Client
}

// Close a client
func (c *Client) Close() {
	c.SSH.Close()
}

func getConfig(h Host) (*ssh.ClientConfig, error) {
	config := ssh.ClientConfig{User: h.GetUserName(),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Minute}
	var err error
	config.Auth, err = h.GetAuth()
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ssh appears to fail with multiple dials
var singleEntry = sync.Mutex{}

// Dial to the host and return an ssh client
func Dial(dest Host) (*Client, error) {
	singleEntry.Lock()
	defer singleEntry.Unlock()
	return dial(dest, map[string]struct{}{})
}

func dial(dest Host, cycle map[string]struct{}) (*Client, error) {
	logger := log.WithField("host", dest.GetID())
	logger.Debugf("dial %+v", dest)
	if _, ok := cycle[dest.GetHostAndPort()]; ok {
		return nil, fmt.Errorf("Cyclic hosts")
	}
	cycle[dest.GetHostAndPort()] = struct{}{}

	var viaCli, client *Client
	var err error
	if v := dest.Via(); v != nil {
		logger.Debugf("Dest %s via %s", dest.GetHostAndPort(), v.GetHostAndPort())
		viaCli, err = dial(v, cycle)
		if err != nil {
			return nil, err
		}
	}

	logger.Debugf("Get client config for %s", dest.GetHostAndPort())
	cfg, err := getConfig(dest)
	if err != nil {
		logger.Errorf("Cannot get client config for %+v: %s", dest, err.Error())
		return nil, err
	}
	logger.Debugf("Dialing %s %s", dest.GetNetwork(), dest.GetHostAndPort())
	if viaCli != nil {
		logger.Debugf("Dialing via bastion")
		conn, err := viaCli.SSH.Dial(dest.GetNetwork(), dest.GetHostAndPort())
		if err != nil {
			log.Debugf("Dial failed for %s %s: %s", dest.GetNetwork(), dest.GetHostAndPort(), err.Error())
			panic(err)
		}
		logger.Debugf("Creating new ssh connection to %s", dest.GetHostAndPort())
		ncc, chans, reqs, err := ssh.NewClientConn(conn, dest.GetHostAndPort(), cfg)
		if err != nil {
			log.Debugf("New connection failed %s: %s", dest.GetHostAndPort(), err.Error())
			panic(err)
		}
		client = &Client{SSH: ssh.NewClient(ncc, chans, reqs)}
	} else {
		logger.Debugf("Dialing directly")
		c, err := ssh.Dial(dest.GetNetwork(), dest.GetHostAndPort(), cfg)
		if err != nil {
			logger.Debugf("Dial failed for %s %s: %s", dest.GetNetwork(), dest.GetHostAndPort(), err.Error())
			panic(err)
		}
		client = &Client{SSH: c}
	}
	return client, nil
}
