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
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/DevFactory/go-tools/pkg/linux/command"
	cmdmock "github.com/DevFactory/go-tools/pkg/linux/command/mock"
	"github.com/DevFactory/go-tools/pkg/nettools"
	nt "github.com/DevFactory/go-tools/pkg/nettools"
	netth "github.com/DevFactory/go-tools/pkg/nettools/testhelpers"
	"github.com/stretchr/testify/assert"
)

const (
	awkIPTablesForNatTest = "sh -c iptables-save | awk -v table=nat -v chain=test " +
		`'$0 ~ "^*"table"$" {in_table=1};$1 ~ "^COMMIT$" {in_table=0};in_table == 1 && $2 ~ "^"chain"$" {print $0}'`
)

func Test_execIPTablesHelper_EnsureChainExists(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		chainName      string
		expectedResult error
		mockInfo       []*cmdmock.ExecInfo
	}{
		{
			name:           "ensure chain when doesn't exist",
			tableName:      "nat",
			chainName:      "test",
			expectedResult: nil,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -n -w 1 -t nat -L test",
					Returned: execResultIPTablesNotFound(),
				},
				{
					Expected: "iptables -w 1 -t nat -N test",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:           "ensure chain when exists",
			tableName:      "nat",
			chainName:      "test",
			expectedResult: nil,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -n -w 1 -t nat -L test",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:           "ensure chain when exists with 1 iptables retry",
			tableName:      "nat",
			chainName:      "test",
			expectedResult: nil,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -n -w 1 -t nat -L test",
					Returned: execResultIPTablesTimeout(),
				},
				{
					Expected: "iptables -n -w 1 -t nat -L test",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:           "ensure chain when exists with 3 iptables retry",
			tableName:      "nat",
			chainName:      "test",
			expectedResult: &exec.ExitError{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -n -w 1 -t nat -L test",
					Returned: execResultIPTablesTimeout(),
				},
				{
					Expected: "iptables -n -w 1 -t nat -L test",
					Returned: execResultIPTablesTimeout(),
				},
				{
					Expected: "iptables -n -w 1 -t nat -L test",
					Returned: execResultIPTablesTimeout(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			err := ipTablesHelper.EnsureChainExists(tt.tableName, tt.chainName)
			assert.Equal(t, tt.expectedResult, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPTablesHelper_EnsureJumpToChainExists(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		baseChain string
		chainName string
		mockInfo  []*cmdmock.ExecInfo
	}{
		{
			name:      "ensure jump to chain when doesn't exist",
			tableName: "nat",
			baseChain: "SN-PREROUTING",
			chainName: "test",
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -C SN-PREROUTING -m comment --comment \"for SNM\" -j test",
					Returned: execResultIPTablesNotFound(),
				},
				{
					Expected: "iptables -w 1 -t nat -A SN-PREROUTING -m comment --comment \"for SNM\" -j test",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:      "ensure jump to chain when exists",
			tableName: "nat",
			baseChain: "SN-PREROUTING",
			chainName: "test",
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -C SN-PREROUTING -m comment --comment \"for SNM\" -j test",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			err := ipTablesHelper.EnsureJumpToChainExists(tt.tableName, tt.chainName, tt.baseChain)
			assert.Nil(t, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPTablesHelper_EnsureExistsInsert(t *testing.T) {
	tests := []struct {
		name     string
		rule     nettools.IPTablesRuleArgs
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name: "ensure rule when doesn't exist",
			rule: nettools.IPTablesRuleArgs{
				Table:     "nat",
				ChainName: "SN-123",
				Selector:  []string{"-p", "tcp", "--dport", "8080"},
				Action:    []string{"DNAT", "--to-destination", "127.0.0.1:9090"},
				Comment:   "for mapping default/map1 [tcp:8080:9090]",
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -C SN-123 -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090",
					Returned: execResultIPTablesNotFound(),
				},
				{
					Expected: "iptables -w 1 -t nat -I SN-123 -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name: "ensure rule when exists",
			rule: nettools.IPTablesRuleArgs{
				Table:     "nat",
				ChainName: "SN-123",
				Selector:  []string{"-p", "tcp", "--dport", "8080"},
				Action:    []string{"DNAT", "--to-destination", "127.0.0.1:9090"},
				Comment:   "for mapping default/map1 [tcp:8080:9090]",
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -C SN-123 -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			err := ipTablesHelper.EnsureExistsInsert(tt.rule)
			assert.Nil(t, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPTablesHelper_Delete(t *testing.T) {
	tests := []struct {
		name     string
		rule     nettools.IPTablesRuleArgs
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name: "delete rule when doesn't exist",
			rule: nettools.IPTablesRuleArgs{
				Table:     "nat",
				ChainName: "SN-123",
				Selector:  []string{"-p", "tcp", "--dport", "8080"},
				Action:    []string{"DNAT", "--to-destination", "127.0.0.1:9090"},
				Comment:   "for mapping default/map1 [tcp:8080:9090]",
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -C SN-123 -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090",
					Returned: execResultIPTablesNotFound(),
				},
			},
		},
		{
			name: "delete rule when exists",
			rule: nettools.IPTablesRuleArgs{
				Table:     "nat",
				ChainName: "SN-123",
				Selector:  []string{"-p", "tcp", "--dport", "8080"},
				Action:    []string{"DNAT", "--to-destination", "127.0.0.1:9090"},
				Comment:   "for mapping default/map1 [tcp:8080:9090]",
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -C SN-123 -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "iptables -w 1 -t nat -D SN-123 -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			err := ipTablesHelper.Delete(tt.rule)
			assert.Nil(t, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPTablesHelper_FlushChain(t *testing.T) {
	tests := []struct {
		name          string
		table         string
		chain         string
		expectedError bool
		mockInfo      []*cmdmock.ExecInfo
	}{
		{
			name:          "flush chain when doesn't exist",
			table:         "nat",
			chain:         "test",
			expectedError: true,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -F test",
					Returned: execResultIPTablesNotFound(),
				},
			},
		},
		{
			name:          "flush chain when exists",
			table:         "nat",
			chain:         "test",
			expectedError: false,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -F test",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			err := ipTablesHelper.FlushChain(tt.table, tt.chain)
			assert.True(t, (err == nil) != tt.expectedError)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPTablesHelper_DeleteChain(t *testing.T) {
	tests := []struct {
		name          string
		table         string
		chain         string
		expectedError bool
		mockInfo      []*cmdmock.ExecInfo
	}{
		{
			name:          "delete chain when doesn't exist",
			table:         "nat",
			chain:         "test",
			expectedError: true,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -X test",
					Returned: execResultIPTablesNotFound(),
				},
			},
		},
		{
			name:          "delete chain when exists",
			table:         "nat",
			chain:         "test",
			expectedError: false,
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: "iptables -w 1 -t nat -X test",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			err := ipTablesHelper.DeleteChain(tt.table, tt.chain)
			assert.True(t, (err == nil) != tt.expectedError)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPTablesHelper_LoadRules(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		chain    string
		expected []*nettools.IPTablesRuleArgs
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:     "parse empty",
			table:    "nat",
			chain:    "test",
			expected: []*nettools.IPTablesRuleArgs{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:  "parse rules",
			table: "nat",
			chain: "test",
			expected: []*nettools.IPTablesRuleArgs{
				{
					Table:     "nat",
					ChainName: "test",
					Selector:  strings.Split("-p tcp --dport 8080", " "),
					Action:    strings.Split("DNAT --to-destination 127.0.0.1:9090", " "),
					Comment:   "for mapping default/map1 [tcp:8080:9090]",
				},
				{
					Table:     "nat",
					ChainName: "test",
					Selector:  strings.Split("-p tcp --dport 3030", " "),
					Action:    strings.Split("DNAT --to-destination 127.0.0.1:2020", " "),
					Comment:   "for mapping default/map2 [tcp:3030:2020]",
				},
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: execResultIPTablesSave2OKRules(),
				},
			},
		},
		{
			name:  "parse rule with no selector",
			table: "nat",
			chain: "test",
			expected: []*nettools.IPTablesRuleArgs{
				{
					Table:     "nat",
					ChainName: "test",
					Selector:  []string{},
					Action:    strings.Split("MARK --set-xmark 0x100000/0x100000", " "),
					Comment:   "mark for masquerade in SNM-POSTROUTING-MASQ",
				},
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: execResultIPTablesSaveNoSelectorRule(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			rules, err := ipTablesHelper.LoadRules(tt.table, tt.chain)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, rules)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPTablesHelper_DeleteByComment(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		chain    string
		comment  string
		expected []*nettools.IPTablesRuleArgs
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:     "delete from empty",
			table:    "nat",
			chain:    "test",
			comment:  "for mapping default/map1 [tcp:8080:9090]",
			expected: []*nettools.IPTablesRuleArgs{},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:    "delete 1 out of 2 rules",
			table:   "nat",
			chain:   "test",
			comment: "for mapping default/map1 [tcp:8080:9090]",
			expected: []*nettools.IPTablesRuleArgs{
				{
					Table:     "nat",
					ChainName: "test",
					Selector:  strings.Split("-p tcp --dport 8080", " "),
					Action:    strings.Split("DNAT --to-destination 127.0.0.1:9090", " "),
					Comment:   "for mapping default/map1 [tcp:8080:9090]",
				},
				{
					Table:     "nat",
					ChainName: "test",
					Selector:  strings.Split("-p tcp --dport 3030", " "),
					Action:    strings.Split("DNAT --to-destination 127.0.0.1:2020", " "),
					Comment:   "for mapping default/map2 [tcp:3030:2020]",
				},
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: execResultIPTablesSave2OKRules(),
				},
				{
					Expected: "iptables -w 1 -t nat -D test -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			err := ipTablesHelper.DeleteByComment(tt.table, tt.chain, tt.comment)
			assert.Nil(t, err)
			execMock.ValidateCallNum()
		})
	}
}

func Test_execIPTablesHelper_EnsureExistsOnlyAppend(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		chain    string
		args     nettools.IPTablesRuleArgs
		mockInfo []*cmdmock.ExecInfo
	}{
		{
			name:  "ensure when exists and nothing to remove",
			table: "nat",
			chain: "test",
			args: nettools.IPTablesRuleArgs{
				Table:     "nat",
				ChainName: "test",
				Selector:  strings.Split("-p tcp --dport 8080", " "),
				Action:    strings.Split("DNAT --to-destination 127.0.0.1:9090", " "),
				Comment:   "for mapping default/map1 [tcp:8080:9090]",
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: execResultIPTablesSave2OKRules(),
				},
			},
		},
		{
			name:  "ensure when doesn't exist and nothing to remove",
			table: "nat",
			chain: "test",
			args: nettools.IPTablesRuleArgs{
				Table:     "nat",
				ChainName: "test",
				Selector:  strings.Split("-p tcp --dport 8081", " "),
				Action:    strings.Split("DNAT --to-destination 127.0.0.1:9091", " "),
				Comment:   "for mapping default/map1 [tcp:8081:9091]",
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: execResultIPTablesSave2OKRules(),
				},
				{
					Expected: "iptables -w 1 -t nat -A test -p tcp --dport 8081 -m comment --comment \"for mapping default/map1 [tcp:8081:9091]\" -j DNAT --to-destination 127.0.0.1:9091",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:  "ensure when exists and some other rules to remove",
			table: "nat",
			chain: "test",
			args: nettools.IPTablesRuleArgs{
				Table:     "nat",
				ChainName: "test",
				Selector:  strings.Split("-p tcp --dport 8080", " "),
				Action:    strings.Split("DNAT --to-destination 127.0.0.1:9090", " "),
				Comment:   "for mapping default/map1 [tcp:8080:9090]",
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: execResultIPTablesSave2RulesSameComment(),
				},
				{
					Expected: "iptables -w 1 -t nat -D test -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 192.168.0.1:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
		{
			name:  "ensure when doesn't exist and some other rules to remove",
			table: "nat",
			chain: "test",
			args: nettools.IPTablesRuleArgs{
				Table:     "nat",
				ChainName: "test",
				Selector:  strings.Split("-p tcp --dport 8080", " "),
				Action:    strings.Split("DNAT --to-destination 127.0.0.10:9090", " "),
				Comment:   "for mapping default/map1 [tcp:8080:9090]",
			},
			mockInfo: []*cmdmock.ExecInfo{
				{
					Expected: awkIPTablesForNatTest,
					Returned: execResultIPTablesSave2RulesSameComment(),
				},
				{
					Expected: "iptables -w 1 -t nat -D test -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "iptables -w 1 -t nat -D test -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 192.168.0.1:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
				{
					Expected: "iptables -w 1 -t nat -A test -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.10:9090",
					Returned: netth.ExecResultOKNoOutput(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := cmdmock.NewMockExecutorFromInfos(t, tt.mockInfo...)
			ipTablesHelper := nt.NewExecIPTablesHelper(execMock, time.Second)
			err := ipTablesHelper.EnsureExistsOnlyAppend(tt.args)
			assert.Nil(t, err)
			execMock.ValidateCallNum()
		})
	}
}

func execResultIPTablesNotFound() *command.ExecResult {
	return &command.ExecResult{
		ExitCode: 1,
		StdErr:   "iptables: No chain/target/match by that name.",
		Err:      &exec.ExitError{},
	}
}

func execResultIPTablesSave2OKRules() *command.ExecResult {
	return &command.ExecResult{
		ExitCode: 0,
		StdOut: "-A test -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090\n" +
			"-A test -p tcp --dport 3030 -m comment --comment \"for mapping default/map2 [tcp:3030:2020]\" -j DNAT --to-destination 127.0.0.1:2020\n",
	}
}

func execResultIPTablesSaveNoSelectorRule() *command.ExecResult {
	return &command.ExecResult{
		ExitCode: 0,
		StdOut:   "-A test -m comment --comment \"mark for masquerade in SNM-POSTROUTING-MASQ\" -j MARK --set-xmark 0x100000/0x100000\n",
	}
}

func execResultIPTablesSave2RulesSameComment() *command.ExecResult {
	return &command.ExecResult{
		ExitCode: 0,
		StdOut: "-A test -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 127.0.0.1:9090\n" +
			"-A test -p tcp --dport 8080 -m comment --comment \"for mapping default/map1 [tcp:8080:9090]\" -j DNAT --to-destination 192.168.0.1:9090\n",
	}
}

func execResultIPTablesTimeout() *command.ExecResult {
	return &command.ExecResult{
		ExitCode: 4,
		StdOut:   "",
		StdErr:   "Another app is currently holding the xtables lock. Perhaps you want to use the -w option?",
		Err:      &exec.ExitError{},
	}
}
