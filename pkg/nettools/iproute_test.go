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
	"testing"

	"github.com/DevFactory/go-tools/pkg/linux/command"
	cmdmock "github.com/DevFactory/go-tools/pkg/linux/command/mock"
	nt "github.com/DevFactory/go-tools/pkg/nettools"
	netmock "github.com/DevFactory/go-tools/pkg/nettools/mocks"
	netth "github.com/DevFactory/go-tools/pkg/nettools/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_execIPRouteHelper_RemoveIPRuleForSourceIP(t *testing.T) {
	tests := []struct {
		name     string
		rule     nt.IPRule
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:     "remove existing rule",
			rule:     nt.NewIPRuleForSourceIP("192.168.3.1", "eth2"),
			mockInfo: netth.GetExecMockResultsForExistingIPRuleRemove(nt.NewIPRuleForSourceIP("192.168.3.1", "eth2")),
		},
		{
			name: "try to remove non existing rule",
			rule: nt.NewIPRuleForSourceIP("192.168.3.11", "eth2"),
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "sh -c ip rule | grep \"from 192\\.168\\.3\\.11 lookup eth2 $\"",
					Returned: netth.ExecResultGrepNotFound(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipRouteHelper := nt.NewExecIPRouteHelper(execMock, nil)
			time, err := ipRouteHelper.RemoveIPRuleForSourceIP(tt.rule)
			assert.Nil(t, err)
			assert.NotZero(t, time)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPRouteHelper_RemoveAllIPRulesForAddressesInSubnet(t *testing.T) {
	tests := []struct {
		name     string
		subnet   net.IPNet
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name: "remove 2 rules",
			subnet: net.IPNet{
				IP:   net.ParseIP("192.168.3.0"),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ip rule",
					Returned: netth.ExecResultExitCodeStdOutput(0,
						"11:	from 10.69.10.210 lookup 10.69.10.211\n"+
							"12:	from 192.168.3.1 lookup eth2\n"+
							"13:	from 192.168.3.2 lookup eth2\n"),
				},
				{
					Expected: "ip rule del from 192.168.3.1 lookup eth2",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "ip rule del from 192.168.3.2 lookup eth2",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name: "not remove any rule",
			subnet: net.IPNet{
				IP:   net.ParseIP("192.168.3.0"),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ip rule",
					Returned: netth.ExecResultExitCodeStdOutput(0, "11:	from 10.69.10.210 lookup 10.69.10.211\n"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			iprouteHelper := nt.NewExecIPRouteHelper(execMock, nil)
			rules, time, err := iprouteHelper.RemoveAllIPRulesForAddressesInSubnet(tt.subnet)
			assert.Nil(t, err)
			assert.NotZero(t, time)
			assert.Len(t, rules, len(tt.mockInfo)-1)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPRouteHelper_EnsureOnlyOneIPRuleExistsForSourceIP(t *testing.T) {
	rule1 := nt.NewIPRuleForSourceIP("192.168.3.1", "eth2")
	tests := []struct {
		name                      string
		rule                      nt.IPRule
		mockInfo                  []*cmdmock.ExecInfo
		expectedRuleToRemoveCount int
	}{
		{
			name: "ensure 1 rule when other exists for the same IP",
			rule: rule1,
			mockInfo: netth.GetExecMockResultsForEnsureOnly1IPRule(rule1, []nt.IPRule{
				nt.NewIPRuleForSourceIP("192.168.3.1", "eth1"),
				nt.NewIPRuleForSourceIP("192.168.3.1", "eth3"),
				nt.NewIPRuleForSourceIP("192.168.3.1", "eth3"),
			}),
			expectedRuleToRemoveCount: 3,
		},
		{
			name:                      "ensure 1 rule when no other rule exists for the same IP",
			rule:                      rule1,
			mockInfo:                  netth.GetExecMockResultsForEnsureOnly1IPRule(rule1, nil),
			expectedRuleToRemoveCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipRouteHelper := nt.NewExecIPRouteHelper(execMock, nil)
			rules, time, err := ipRouteHelper.EnsureOnlyOneIPRuleExistsForSourceIP(tt.rule)
			assert.Nil(t, err)
			assert.NotZero(t, time)
			assert.Len(t, rules, tt.expectedRuleToRemoveCount)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPRouteHelper_EnsureOnlyOneIPRuleExistsForFwMark(t *testing.T) {
	rule1 := nt.NewIPRuleForFwMark(0x10, "eth2")
	tests := []struct {
		name                      string
		rule                      nt.IPRule
		mockInfo                  []*cmdmock.ExecInfo
		expectedRuleToRemoveCount int
	}{
		{
			name: "ensure 1 rule when other exists for the same fwmark",
			rule: rule1,
			mockInfo: netth.GetExecMockResultsForEnsureOnly1IPRule(rule1, []nt.IPRule{
				nt.NewIPRuleForFwMark(0x10, "eth1"),
				nt.NewIPRuleForFwMark(0x10, "eth3"),
				nt.NewIPRuleForFwMark(0x10, "eth3"),
			}),
			expectedRuleToRemoveCount: 3,
		},
		{
			name:                      "ensure 1 rule when no other rule exists for the same IP",
			rule:                      rule1,
			mockInfo:                  netth.GetExecMockResultsForEnsureOnly1IPRule(rule1, nil),
			expectedRuleToRemoveCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipRouteHelper := nt.NewExecIPRouteHelper(execMock, nil)
			rules, time, err := ipRouteHelper.EnsureOnlyOneIPRuleExistsForFwMark(tt.rule)
			assert.Nil(t, err)
			assert.NotZero(t, time)
			assert.Len(t, rules, tt.expectedRuleToRemoveCount)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPRouteHelper_InitializeRoutingTablesPerInterface(t *testing.T) {
	mockInterfaces := netmock.MockInterfacesEth
	mockInfo := netth.GetExecInfosForIPRouteInterfaceInit(mockInterfaces)
	tests := []struct {
		name     string
		ifaces   []nt.Interface
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:     "check create 2 eth interfaces",
			ifaces:   mockInterfaces,
			mockInfo: mockInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ioOp := netth.GetMockIOOpProviderWithEmptyRTTablesFile([]byte{}, len(tt.ifaces))
			ipRouteHelper := nt.NewExecIPRouteHelper(execMock, ioOp)
			time, err := ipRouteHelper.InitializeRoutingTablesPerInterface(tt.ifaces)
			assert.Nil(t, err)
			assert.NotZero(t, time)
			execMock.ValidateCallNum()
			ioOp.AssertExpectations(t)
		})
	}
}

func Test_execIPRouteHelper_EnsureRoutes(t *testing.T) {
	entries := []nt.IPRouteEntry{
		{
			TableName:    "eth1",
			TargetPrefix: "10.10.10.0/24",
			Mode:         "dev",
			Gateway:      "eth1",
			Options:      "scope local",
		},
		{
			TableName:    "eth1",
			TargetPrefix: "default",
			Mode:         "via",
			Gateway:      "10.10.10.1",
			Options:      "dev eth1",
		},
	}
	tests := []struct {
		name     string
		entries  []nt.IPRouteEntry
		mockInfo []*cmdmock.ExecInfo
		wantErr  bool
	}{
		{
			name:    "Ensures 2 rules when table empty",
			entries: entries,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ip route show tab eth1",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "ip route add 10.10.10.0/24 dev eth1 scope local tab eth1",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "ip route add default via 10.10.10.1 dev eth1 tab eth1",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
			wantErr: false,
		},
		{
			name:    "Ensures 2 rules when table has 1 of them",
			entries: entries,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ip route show tab eth1",
					Returned: &command.ExecResult{
						StdOut: "default via 10.10.10.1 dev eth1\n",
					},
				},
				{
					Expected: "ip route add 10.10.10.0/24 dev eth1 scope local tab eth1",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
			wantErr: false,
		},
		{
			name:    "Ensures 2 rules when table has 2 of them",
			entries: entries,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ip route show tab eth1",
					Returned: &command.ExecResult{
						StdOut: "10.10.10.0/24 dev eth1 scope link\ndefault via 10.10.10.1 dev eth1\n",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Ensures 2 rules when table has others",
			entries: entries,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "ip route show tab eth1",
					Returned: &command.ExecResult{
						StdOut: "10.10.20.0/24 dev eth1 scope link\ndefault via 10.10.20.1 dev eth1\n",
					},
				},
				{
					Expected: "ip route add 10.10.10.0/24 dev eth1 scope local tab eth1",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "ip route add default via 10.10.10.1 dev eth1 tab eth1",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipRouteHelper := nt.NewExecIPRouteHelper(execMock, nil)

			_, err := ipRouteHelper.EnsureRoutes(tt.entries)
			if (err != nil) != tt.wantErr {
				t.Errorf("execIPRouteHelper.EnsureRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
