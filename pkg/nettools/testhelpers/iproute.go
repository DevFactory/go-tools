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

package testhelpers

import (
	"fmt"
	"strings"

	"github.com/DevFactory/go-tools/pkg/linux/command"
	cmdmock "github.com/DevFactory/go-tools/pkg/linux/command/mock"
	nt "github.com/DevFactory/go-tools/pkg/nettools"
	"github.com/DevFactory/go-tools/pkg/nettools/mocks"
	"github.com/stretchr/testify/mock"
)

// GetExecInfosForIPRouteInterfaceInit returns mock ExecInfo slice for commands
// executed during initialization of interfaces
func GetExecInfosForIPRouteInterfaceInit(interfaces []nt.Interface) []*cmdmock.ExecInfo {
	mockInfo := []*cmdmock.ExecInfo{}
	for _, iface := range interfaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			rule := nt.IPRule{
				No:             0,
				SourceIP:       addr.IP,
				RouteTableName: iface.GetName(),
			}
			mockInfo = append(mockInfo, GetExecMockResultsForEnsureOnly1IPRule(rule, nil)...)
		}
	}
	return mockInfo
}

// GetMockIOOpProviderWithEmptyRTTablesFile returns mock IOOpProvider that mocks an empty rttables file
func GetMockIOOpProviderWithEmptyRTTablesFile(rtTablesReadResult []byte, times int, 
	withAppend bool) *mocks.MockSimpleFileOperator {
	ioOp := new(mocks.MockSimpleFileOperator)
	ioOp.On("ReadFile", nt.RTTablesFilename).Return(rtTablesReadResult, nil).Times(times)
	if withAppend {
		ioOp.On("AppendToFile", nt.RTTablesFilename, mock.Anything).Return(nil).Times(times)
	}
	return ioOp
}

func GetExecMockResultsForEnsureOnly1IPRule(rule nt.IPRule, otherRules []nt.IPRule) []*cmdmock.ExecInfo {
	checkExpected := ""
	if rule.SourceIP != nil {
		grepIP, grepTableName := rule.GrepEscaped()
		checkExpected = fmt.Sprintf("sh -c ip rule | grep -v \"from %s lookup %s $\" | grep \"from %s lookup \"",
			grepIP, grepTableName, grepIP)
	} else {
		checkExpected = fmt.Sprintf("sh -c ip rule | grep -v \"fwmark %s lookup %s $\" | grep \"fwmark %s lookup \"",
			rule.FwMarkHexed(), rule.RouteTableName, rule.FwMarkHexed())
	}
	checkReturned := ""
	for _, other := range otherRules {
		checkReturned += other.ToString() + "\n"
	}
	res := []*cmdmock.ExecInfo{
		{
			Expected: checkExpected,
			Returned: ExecResultGrepFound(checkReturned),
		},
	}
	for _, other := range otherRules {
		res = append(res, GetExecMockResultsForExistingIPRuleRemove(other)...)
	}
	res = append(res, getExecMockResultsForNonExistingIPRuleAdd(rule)...)
	return res
}

func getExecMockResultsForIPRuleOperation(rule nt.IPRule, operation string, grepResult *command.ExecResult,
	key, arg string) []*cmdmock.ExecInfo {

	grepArg := strings.Replace(arg, ".", `\.`, 3)
	res := []*cmdmock.ExecInfo{}
	res = append(res, &cmdmock.ExecInfo{
		Expected: fmt.Sprintf("sh -c ip rule | grep \"%s %s lookup %s $\"", key, grepArg, rule.RouteTableName),
		Returned: grepResult,
	})
	res = append(res, &cmdmock.ExecInfo{
		Expected: fmt.Sprintf("ip rule %s %s %s lookup %s", operation, key, arg, rule.RouteTableName),
		Returned: ExecResultOKNoOutput(),
	})
	return res
}

func GetExecMockResultsForExistingIPRuleRemove(rule nt.IPRule) []*cmdmock.ExecInfo {
	var args *command.ExecResult
	key := ""
	arg := ""

	if rule.SourceIP != nil {
		args = ExecResultGrepFound(fmt.Sprintf("11:	from %s lookup %s ", rule.SourceIP.String(), rule.RouteTableName))
		key = "from"
		arg = rule.SourceIP.String()
	} else {
		args = ExecResultGrepFound(fmt.Sprintf("11:	from all fwmark %s lookup %s ", rule.FwMarkHexed(), rule.RouteTableName))
		key = "fwmark"
		arg = rule.FwMarkHexed()
	}
	return getExecMockResultsForIPRuleOperation(rule, "del", args, key, arg)
}

func getExecMockResultsForNonExistingIPRuleAdd(rule nt.IPRule) []*cmdmock.ExecInfo {
	key := ""
	arg := ""
	if rule.SourceIP != nil {
		key = "from"
		arg = rule.SourceIP.String()
	} else {
		key = "fwmark"
		arg = rule.FwMarkHexed()
	}
	return getExecMockResultsForIPRuleOperation(rule, "add", ExecResultGrepNotFound(), key, arg)
}
