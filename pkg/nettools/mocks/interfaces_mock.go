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

package mocks

import (
	"net"
	"time"

	nt "github.com/DevFactory/go-tools/pkg/nettools"
)

// NewMockInterfaceProvider creates a new InterfaceProvider that delivers
// information about local interfaces from the in-code constant
func NewMockInterfaceProvider(managedInterfacesRegexp string, autoRefresh bool) (nt.InterfaceProvider,
	chan time.Time, error) {
	ch := make(chan time.Time)
	ip, err := nt.NewChanInterfaceProvider(ch, &MockInterfaceLister{}, managedInterfacesRegexp,
		autoRefresh)
	return ip, ch, err
}

var MockInterfaces = append(MockInterfacesEth, MockInterfacesOther...)

var MockInterfacesEth = []nt.Interface{
	&mockInterface{
		Name: "eth0",
		Addresses: []net.IPNet{
			{
				IP:   net.ParseIP("192.168.1.1"),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
		},
		Index: 1,
	},
	&mockInterface{
		Name: "eth2",
		Addresses: []net.IPNet{
			{
				IP:   net.ParseIP("10.10.10.10"),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
			{
				IP:   net.ParseIP("10.10.10.11"),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
		},
		Index: 2,
	},
}

var MockInterfacesOther = []nt.Interface{
	&mockInterface{
		Name: "lo",
		Addresses: []net.IPNet{
			{
				IP:   net.ParseIP("127.0.0.1"),
				Mask: net.IPv4Mask(255, 255, 255, 255),
			},
		},
		Index: 3,
	},
	&mockInterface{
		Name: "wifi1",
		Addresses: []net.IPNet{
			{
				IP:   net.ParseIP("10.50.10.10"),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
		},
		Index: 4,
	},
}

type MockInterfaceLister struct {
}

func (m *MockInterfaceLister) GetInterfaces() ([]nt.Interface, error) {
	return MockInterfaces, nil
}

type mockInterface struct {
	Name      string
	Addresses []net.IPNet
	Index     int
}

func (n *mockInterface) Addrs() ([]net.IPNet, error) {
	return n.Addresses, nil
}

func (n *mockInterface) GetName() string {
	return n.Name
}

func (n *mockInterface) GetIndex() int {
	return n.Index
}

func (n *mockInterface) Copy() nt.Interface {
	addrs := make([]net.IPNet, 0, len(n.Addresses))
	for _, addr := range n.Addresses {
		addrs = append(addrs, addr)
	}
	return &mockInterface{
		Addresses: addrs,
		Name:      n.Name,
	}
}
