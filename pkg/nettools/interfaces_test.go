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

package nettools_test

import (
	"net"
	"reflect"
	"regexp"
	"testing"
	"time"

	nt "github.com/DevFactory/go-tools/pkg/nettools"
	"github.com/DevFactory/go-tools/pkg/nettools/mocks"
	"github.com/stretchr/testify/assert"
)

var refreshTime = time.Microsecond

func Test_netInterfaceProvider_GetInterfaceForLocalIP_IsLocalIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      net.IP
		want    nt.Interface
		wantErr bool
	}{
		{
			name:    "finds interface for matching interface and IP",
			ip:      net.ParseIP("192.168.1.1"),
			want:    mocks.MockInterfacesEth[0],
			wantErr: false,
		},
		{
			name:    "doesn't find interface for interface not matching regexp",
			ip:      net.ParseIP("127.0.0.1"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "doesn't find interface for not existing IP",
			ip:      net.ParseIP("1.2.3.4"),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		nip, _, _ := mocks.NewMockInterfaceProvider("^eth[0-9]+$", true)
		t.Run(tt.name, func(t *testing.T) {
			got, err := nip.GetInterfaceForLocalIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("netInterfaceProvider.GetInterfaceForLocalIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("netInterfaceProvider.GetInterfaceForLocalIP() = %v, want %v", got, tt.want)
			}
			isLocal := nip.IsLocalIP(tt.ip)
			if isLocal == tt.wantErr {
				t.Errorf("netInterfaceProvider.IsLocalIP() returned = %v, expected %v", isLocal, !tt.wantErr)
				return
			}
		})
	}
}

func Test_Refreshing_Interfaces(t *testing.T) {
	regExpString := "^eth[0-9]+$"
	regExp, _ := regexp.Compile(regExpString)
	nip, sigChan, _ := mocks.NewMockInterfaceProvider(regExpString, true)

	nip.StartRefreshing()
	sigChan <- time.Now()
	for nip.GetRefreshCount() < 1 {
		time.Sleep(time.Millisecond)
	}
	nip.StopRefreshing()
	ifaces, err := nip.Interfaces()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ifaces))
	for ii, iface := range ifaces {
		assert.True(t, regExp.MatchString(iface.GetName()))

		addrs, err := iface.Addrs()
		assert.Nil(t, err)

		expectedAddrs, err := mocks.MockInterfacesEth[ii].Addrs()
		assert.Equal(t, len(expectedAddrs), len(addrs))
		for i := 0; i < len(addrs); i++ {
			assert.Equal(t, expectedAddrs[i], addrs[i])
		}
	}
	assert.True(t, nip.GetRefreshCount() > 0)
}
