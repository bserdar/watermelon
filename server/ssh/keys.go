package ssh

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

// RawPrivateKey contains the data for a private key. It can be from a
// file, or direct from the data
type RawPrivateKey struct {
	sync.RWMutex
	Name       string
	File       string
	PEMData    []byte
	Passphrase string

	pk interface{}
}

// GetPrivateKey reads the key from a file, asks passphrase if
// necessary, and returns a private key. Once the private key is
// retrieved, the same instance is returned, so keep references to
// this instance, don't copy
func (x *RawPrivateKey) GetPrivateKey() (interface{}, error) {
	x.RLock()
	if x.pk != nil {
		x.RUnlock()
		return x.pk, nil
	}
	x.RUnlock()
	x.Lock()
	defer x.Unlock()
	if x.PEMData == nil {
		// No data, load from file
		if len(x.File) > 0 {
			key, err := ioutil.ReadFile(x.File)
			if err != nil {
				return nil, fmt.Errorf("Cannot read private key %s: %s", x.File, err)
			}
			x.PEMData = key
		} else {
			return nil, fmt.Errorf("No private key can be loaded")
		}
	}
	if x.PEMData != nil {
		pk, err := ssh.ParseRawPrivateKey(x.PEMData)
		if err != nil {
			if strings.Contains(err.Error(), "encrypted") {

				if len(x.Passphrase) == 0 {
					var prompt string
					if len(x.Name) > 0 {
						prompt = fmt.Sprintf("Enter passphrase for %s: ", x.Name)
					} else {
						prompt = "Enter passphrase: "
					}
					x.Passphrase = AskPassword(prompt)
				}

				pk, err = ssh.ParseRawPrivateKeyWithPassphrase(x.PEMData, []byte(x.Passphrase))
			}
		}
		x.pk = pk
		return x.pk, err
	}

	return nil, fmt.Errorf("No private key")
}

// GetSigner returns a signer from the private key
func (x *RawPrivateKey) GetSigner() (ssh.Signer, error) {
	pk, err := x.GetPrivateKey()
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(pk)
}
