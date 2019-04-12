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
	"fmt"
	"net"
	"os/exec"
	"testing"

	"github.com/DevFactory/go-tools/pkg/linux/command"
	cmdmock "github.com/DevFactory/go-tools/pkg/linux/command/mock"
	"github.com/DevFactory/go-tools/pkg/nettools"
	nt "github.com/DevFactory/go-tools/pkg/nettools"
	netth "github.com/DevFactory/go-tools/pkg/nettools/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_execIPSetHelper_EnsureSetExists(t *testing.T) {
	tests := []struct {
		name     string
		setName  string
		setType  string
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:    "create a new set",
			setName: "12341234abc",
			setType: "hash:ip",
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ipset create 12341234abc hash:ip comment counters -exist",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipSetHelper := nt.NewExecIPSetHelper(execMock)
			err := ipSetHelper.EnsureSetExists(tt.setName, tt.setType)
			assert.Nil(t, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPSetHelper_DeleteSet(t *testing.T) {
	tests := []struct {
		name     string
		setName  string
		mockInfo []*cmdmock.ExecInfo
		err      error
	}{
		{
			name:    "delete existing ipset",
			setName: "12341234abc",
			err:     nil,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ipset destroy 12341234abc",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:    "delete non existing ipset",
			setName: "12341234abc",
			err: &exec.ExitError{
				Stderr: []byte("ipset v6.34: The set with the given name does not exist"),
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ipset destroy 12341234abc",
					Returned: execResultIpsetNotFound(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipSetHelper := nt.NewExecIPSetHelper(execMock)
			err := ipSetHelper.DeleteSet(tt.setName)
			assert.Equal(t, tt.err, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPSetHelper_GetIPs(t *testing.T) {
	tests := []struct {
		name     string
		setName  string
		err      error
		expected []net.IP
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:     "get from existing empty set",
			setName:  "12341234abc",
			err:      nil,
			expected: []net.IP{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:    "get from non existing set",
			setName: "12341234abc",
			err: &exec.ExitError{
				Stderr: []byte("ipset v6.34: The set with the given name does not exist"),
			},
			expected: []net.IP{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: execResultIpsetNotFound(),
				},
			},
		},
		{
			name:     "get from existing non empty set",
			setName:  "12341234abc",
			err:      nil,
			expected: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2")},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: execResultIpsetIPs(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipSetHelper := nt.NewExecIPSetHelper(execMock)
			ips, err := ipSetHelper.GetIPs(tt.setName)
			assert.Equal(t, tt.expected, ips)
			assert.Equal(t, tt.err, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPSetHelper_EnsureSetHasOnly(t *testing.T) {
	tests := []struct {
		name      string
		setName   string
		err       error
		addresses []net.IP
		mockInfo  []*cmdmock.ExecInfo
	}{
		{
			name:      "sync empty ipset with empty required set",
			setName:   "12341234abc",
			err:       nil,
			addresses: []net.IP{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:      "sync empty ipset with non empty required set",
			setName:   "12341234abc",
			err:       nil,
			addresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2")},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "ipset add 12341234abc 127.0.0.1",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "ipset add 12341234abc 127.0.0.2",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:      "sync non empty ipset with empty required set",
			setName:   "12341234abc",
			err:       nil,
			addresses: []net.IP{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: execResultIpsetIPs(),
				},
				{
					Expected: "ipset del 12341234abc 127.0.0.1",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "ipset del 12341234abc 127.0.0.2",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:      "sync non empty ipset with non empty required set",
			setName:   "12341234abc",
			err:       nil,
			addresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.3")},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: execResultIpsetIPs(),
				},
				{
					Expected: "ipset add 12341234abc 127.0.0.3",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "ipset del 12341234abc 127.0.0.2",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipSetHelper := nt.NewExecIPSetHelper(execMock)
			err := ipSetHelper.EnsureSetHasOnly(tt.setName, tt.addresses)
			assert.Equal(t, tt.err, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPSetHelper_GetNetPorts(t *testing.T) {
	np1, np2 := getSampleNetPorts()
	tests := []struct {
		name     string
		setName  string
		err      error
		expected []nt.NetPort
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:     "get from existing empty set",
			setName:  "12341234abc",
			err:      nil,
			expected: []nt.NetPort{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:    "get from non existing set",
			setName: "12341234abc",
			err: &exec.ExitError{
				Stderr: []byte("ipset v6.34: The set with the given name does not exist"),
			},
			expected: []nt.NetPort{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: execResultIpsetNotFound(),
				},
			},
		},
		{
			name:     "get from existing non empty set",
			setName:  "12341234abc",
			err:      nil,
			expected: []nt.NetPort{np1, np2},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: execResultIpsetNetPorts(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipSetHelper := nt.NewExecIPSetHelper(execMock)
			nps, err := ipSetHelper.GetNetPorts(tt.setName)
			assert.Equal(t, tt.expected, nps)
			assert.Equal(t, tt.err, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPSetHelper_EnsureSetHasOnlyNetPort(t *testing.T) {
	np1, np2 := getSampleNetPorts()
	np3 := getDifferentSampleNetPorts()
	tests := []struct {
		name     string
		setName  string
		err      error
		expected []nt.NetPort
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:     "sync empty ipset with empty required set",
			setName:  "12341234abc",
			err:      nil,
			expected: []nt.NetPort{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:     "sync empty ipset with non empty required set",
			setName:  "12341234abc",
			err:      nil,
			expected: []nt.NetPort{np1, np2},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: fmt.Sprintf("ipset add 12341234abc %s", np1),
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: fmt.Sprintf("ipset add 12341234abc %s", np2),
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:     "sync non empty ipset with empty required set",
			setName:  "12341234abc",
			err:      nil,
			expected: []nt.NetPort{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: execResultIpsetNetPorts(),
				},
				{
					Expected: fmt.Sprintf("ipset del 12341234abc %s", np1),
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: fmt.Sprintf("ipset del 12341234abc %s", np2),
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:     "sync non empty ipset with non empty required set",
			setName:  "12341234abc",
			err:      nil,
			expected: []nt.NetPort{np2, np3},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: fmt.Sprintf("sh -c %s", fmt.Sprintf(nettools.IPSetListWithAwk, "12341234abc")),
					Returned: execResultIpsetNetPorts(),
				},
				{
					Expected: fmt.Sprintf("ipset add 12341234abc %s", np3),
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: fmt.Sprintf("ipset del 12341234abc %s", np1),
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipSetHelper := nt.NewExecIPSetHelper(execMock)
			err := ipSetHelper.EnsureSetHasOnlyNetPort(tt.setName, tt.expected)
			assert.Equal(t, tt.err, err)
			execMock.ValidateCallNum()
		})
	}
}
func execResultIpsetNotFound() *command.ExecResult {
	return &command.ExecResult{
		ExitCode: 1,
		StdErr:   "ipset v6.34: The set with the given name does not exist",
		Err: &exec.ExitError{
			Stderr: []byte("ipset v6.34: The set with the given name does not exist"),
		},
	}
}

func execResultIpsetIPs() *command.ExecResult {
	return &command.ExecResult{
		StdOut: "127.0.0.1\n127.0.0.2\n",
	}
}

func execResultIpsetNetPorts() *command.ExecResult {
	np1, np2 := getSampleNetPorts()
	return &command.ExecResult{
		StdOut: fmt.Sprintf("%s\n%s\n", np1.String(), np2.String()),
	}
}

func getSampleNetPorts() (nt.NetPort, nt.NetPort) {
	_, netAddr1, _ := net.ParseCIDR("10.10.0.0/24")
	np1 := nt.NetPort{
		Net:      *netAddr1,
		Port:     80,
		Protocol: nt.TCP,
	}
	_, netAddr2, _ := net.ParseCIDR("10.20.0.0/24")
	np2 := nt.NetPort{
		Net:      *netAddr2,
		Port:     8080,
		Protocol: nt.UDP,
	}
	return np1, np2
}

func getDifferentSampleNetPorts() nt.NetPort {
	_, netAddr2, _ := net.ParseCIDR("10.120.0.0/24")
	np := nt.NetPort{
		Net:      *netAddr2,
		Port:     8080,
		Protocol: nt.UDP,
	}
	return np
}
