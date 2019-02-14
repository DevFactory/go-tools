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

package net_test

import (
	"net"
	"reflect"
	"testing"

	enet "github.com/DevFactory/go-tools/pkg/extensions/net"
	"github.com/DevFactory/go-tools/pkg/extensions/net/mocks"
)

func TestIPNetFromAddrWithMask(t *testing.T) {
	tests := []struct {
		name    string
		addr    net.Addr
		want    net.IPNet
		wantErr bool
	}{
		{
			name:    "parse correct addr 127.0.0.1/32",
			addr:    &mocks.MockAddress{AddressStr: "127.0.0.1/32", NetworkStr: "ip"},
			want:    net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.IPv4Mask(255, 255, 255, 255)},
			wantErr: false,
		},
		{
			name:    "parse correct addr 192.168.1.1",
			addr:    &mocks.MockAddress{AddressStr: "192.168.1.1", NetworkStr: "ip"},
			want:    net.IPNet{IP: net.ParseIP("192.168.1.1"), Mask: net.IPv4Mask(255, 255, 255, 255)},
			wantErr: false,
		},
		{
			name:    "parse correct addr 192.168.10.0/24",
			addr:    &mocks.MockAddress{AddressStr: "192.168.10.0/24", NetworkStr: "ip"},
			want:    net.IPNet{IP: net.ParseIP("192.168.10.0"), Mask: net.IPv4Mask(255, 255, 255, 0)},
			wantErr: false,
		},
		{
			name:    "parse incorrect addr 292.168.10.0/24",
			addr:    &mocks.MockAddress{AddressStr: "292.168.10.0/24", NetworkStr: "ip"},
			want:    net.IPNet{},
			wantErr: true,
		},
		{
			name:    "parse incorrect mask 192.168.10.0/33",
			addr:    &mocks.MockAddress{AddressStr: "192.168.10.0/33", NetworkStr: "ip"},
			want:    net.IPNet{},
			wantErr: true,
		},
		{
			name:    "parse incorrect mask 192.168.10.0/A",
			addr:    &mocks.MockAddress{AddressStr: "192.168.10.0/A", NetworkStr: "ip"},
			want:    net.IPNet{},
			wantErr: true,
		},
		{
			name:    "parse incorrect CIDR 192.168.10.0/32/32",
			addr:    &mocks.MockAddress{AddressStr: "192.168.10.0/32/32", NetworkStr: "ip"},
			want:    net.IPNet{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := enet.IPNetFromAddrWithMask(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("IPNetFromAddrWithMask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IPNetFromAddrWithMask() = %v, want %v", got, tt.want)
			}
		})
	}
}
