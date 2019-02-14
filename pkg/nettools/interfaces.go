/* Copyright 2019 DevFactory FZ LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package nettools

import (
	"fmt"
	"net"
	"regexp"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	enet "github.com/DevFactory/go-tools/pkg/extensions/net"
	etime "github.com/DevFactory/go-tools/pkg/extensions/time"
)

// Interface provides addresses and name of a single local network interface
//
// Addrs() gets a slice with all addresses assigned to this interface
//
// GetName() returns name of the interfaces, as presented in the operating
// system
//
// Copy() returns a full deep copy of the Interface object
type Interface interface {
	GetIndex() int
	Addrs() ([]net.IPNet, error)
	GetName() string
	Copy() Interface
}

// InterfaceProvider delivers information about local interfaces and assigned IP
// addresses. It also periodically refreshes this information in the background
// to stay in sync with changes in the Operating System.
//
// Interfaces() lists all discovered networking interfaces
//
// IsLocalIP() checks if the ip passed as parameter is assigned to any discovered
// interface
//
// GetInterfaceForLocalIP() returns Interface which has ip assigned. If no such
// interface is found, error must be set.
type InterfaceProvider interface {
	etime.Refresher
	Interfaces() ([]Interface, error)
	IsLocalIP(ip net.IP) bool
	GetInterfaceForLocalIP(ip net.IP) (Interface, error)
}

// InterfaceLister lists all network interfaces in the Operating System, together with
// their assigned IP addresses
type InterfaceLister interface {
	GetInterfaces() ([]Interface, error)
}

// net package based implementation
type netInterfaceProvider struct {
	etime.Refresher
	ifaces          []Interface
	interfaceLister InterfaceLister
	ifaceRegexp     *regexp.Regexp
	ifacePadlock    *sync.RWMutex
}

// NewInterfaceProvider creates a new InterfaceProvider that delivers
// information about local interfaces matching specific Regexp and with
// optional auto-refresh
func NewInterfaceProvider(lister InterfaceLister, managedInterfacesRegexp string,
	autoRefresh bool, autoRefreshPeriod time.Duration) (InterfaceProvider, error) {
	res, err := newInterfaceProviderWithNoRefresher(lister, managedInterfacesRegexp)
	if err != nil {
		return nil, err
	}
	res.Refresher = etime.NewTickerRefresher(res.reloadInterfaces, true, autoRefresh, autoRefreshPeriod)
	return res, nil
}

// NewChanInterfaceProvider creates a new InterfaceProvider that delivers
// information about local interfaces matching specific Regexp and with
// optional auto-refresh and refreshes this information every time a message
// is received on the update chan
func NewChanInterfaceProvider(updateChan chan time.Time, lister InterfaceLister,
	managedInterfacesRegexp string, autoRefresh bool) (InterfaceProvider, error) {
	res, err := newInterfaceProviderWithNoRefresher(lister, managedInterfacesRegexp)
	if err != nil {
		return nil, err
	}
	res.Refresher = etime.NewRefresher(updateChan, res.reloadInterfaces, true, autoRefresh)
	return res, nil
}

func newInterfaceProviderWithNoRefresher(lister InterfaceLister,
	managedInterfacesRegexp string) (*netInterfaceProvider, error) {
	ifaceRegExp, err := regexp.Compile(managedInterfacesRegexp)
	if err != nil {
		return nil, err
	}
	res := &netInterfaceProvider{
		ifaces:          []Interface{},
		interfaceLister: lister,
		ifaceRegexp:     ifaceRegExp,
		ifacePadlock:    &sync.RWMutex{},
	}
	return res, nil
}

// NewNetInterfaceProvider creates a new InterfaceProvider that delivers
// information about local interfaces in the real Operating System using
// golang's net standard package
func NewNetInterfaceProvider(managedInterfacesRegexp string,
	autoRefresh bool, autoRefreshPeriod time.Duration) (InterfaceProvider, error) {
	return NewInterfaceProvider(&netInterfaceLister{}, managedInterfacesRegexp,
		autoRefresh, autoRefreshPeriod)
}

func (r *netInterfaceProvider) Interfaces() ([]Interface, error) {
	r.ifacePadlock.RLock()
	defer r.ifacePadlock.RUnlock()
	res := make([]Interface, len(r.ifaces))
	for i := 0; i < len(r.ifaces); i++ {
		res[i] = r.ifaces[i].Copy()
	}
	return res, nil
}

func (r *netInterfaceProvider) reloadInterfaces() {
	ifaces, err := FilteredInterfaces(r.ifaceRegexp, r.interfaceLister)
	if err != nil {
		log.Errorf("Error loading interface information from the operating system: %v",
			err)
	}
	log.Debugf("InterfaceProvider found %d network interfaces configured for use", len(ifaces))
	r.ifacePadlock.Lock()
	defer r.ifacePadlock.Unlock()
	r.ifaces = ifaces
}

func (r *netInterfaceProvider) IsLocalIP(ip net.IP) bool {
	_, err := r.GetInterfaceForLocalIP(ip)
	return err == nil
}

func (r *netInterfaceProvider) GetInterfaceForLocalIP(ip net.IP) (Interface, error) {
	var res Interface
	r.ifacePadlock.RLock()
	defer r.ifacePadlock.RUnlock()
	for _, iface := range r.ifaces {
		addresses, _ := iface.Addrs()
		for _, addr := range addresses {
			if addr.IP.Equal(ip) {
				res = iface
			}
		}
	}
	if res == nil {
		return nil, fmt.Errorf("Local interface for IP %v not found", ip.String())
	}
	return res, nil
}

type netInterface struct {
	iface net.Interface
}

func fromNetInterface(netIface net.Interface) Interface {
	return &netInterface{netIface}
}

func (n *netInterface) Addrs() ([]net.IPNet, error) {
	addrs, err := n.iface.Addrs()
	if err != nil {
		return nil, err
	}
	res := make([]net.IPNet, len(addrs))
	for i := 0; i < len(addrs); i++ {
		a, _, _ := net.ParseCIDR(addrs[i].String())
		if a.To4() == nil {
			continue
		}
		ipnet, err := enet.IPNetFromAddrWithMask(addrs[i])
		if err != nil {
			return nil, err
		}
		res[i] = ipnet
	}
	return res, nil
}

func (n *netInterface) GetName() string {
	return n.iface.Name
}

func (n *netInterface) Copy() Interface {
	return &netInterface{iface: n.iface}
}

func (n *netInterface) GetIndex() int {
	return n.iface.Index
}

type netInterfaceLister struct{}

func (n *netInterfaceLister) GetInterfaces() ([]Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	log.Debugf("InterfaceLister found %d network interfaces in total", len(ifaces))
	res := make([]Interface, len(ifaces))
	for i, iface := range ifaces {
		res[i] = fromNetInterface(iface)
	}
	return res, nil
}
