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
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/DevFactory/go-tools/pkg/linux/command"
)

// IPRule represents a single ip routing rule based on source IP and selecting a named routing table.
type IPRule struct {
	No             int
	FwMark         int
	SourceIP       net.IP
	RouteTableName string
}

// Copy returns a copy of IPRule using new memory.
func (r *IPRule) Copy() IPRule {
	return IPRule{
		No:             r.No,
		SourceIP:       r.SourceIP,
		RouteTableName: r.RouteTableName,
		FwMark:         r.FwMark,
	}
}

// ToString return a string representation of the rule, as used in "ip" command
func (r *IPRule) ToString() string {
	if r.SourceIP != nil {
		return fmt.Sprintf("%d:\tfrom %s lookup %s ", r.No, r.SourceIP, r.RouteTableName)
	}
	return fmt.Sprintf("%d:\tfrom all fwmark %s lookup %s ", r.No, r.FwMarkHexed(), r.RouteTableName)
}

// GrepEscaped returns IP address and route table name with "." escaped for use with `grep` command
func (r *IPRule) GrepEscaped() (string, string) {
	return strings.Replace(r.SourceIP.String(), ".", `\.`, 3), strings.Replace(r.RouteTableName, ".", `\.`, -1)
}

// FwMarkHexed returns string representing FWMark in hexadecimal format
func (r *IPRule) FwMarkHexed() string {
	hex := fmt.Sprintf("0x%x", r.FwMark)
	return hex
}

// NewIPRuleForSourceIP creates a new IPRule with source IP and with default No of 0
func NewIPRuleForSourceIP(cidrAddress, tableName string) IPRule {
	return IPRule{
		SourceIP:       net.ParseIP(cidrAddress),
		RouteTableName: tableName,
	}
}

// NewIPRuleForFwMark creates a new IPRule with source IP and with default No of 0
func NewIPRuleForFwMark(mark int, tableName string) IPRule {
	return IPRule{
		FwMark:         mark,
		RouteTableName: tableName,
	}
}

// IPRouteEntry represents one routing information entry in a configured route table
type IPRouteEntry struct {
	TableName, TargetPrefix, Mode, Gateway, Options string
}

func (e IPRouteEntry) String() string {
	return fmt.Sprintf("%s: %s %s %s %s", e.TableName, e.TargetPrefix, e.Mode, e.Gateway, e.Options)
}

/*
IPRouteHelper implements various iproute2 tool related actions.

InitializeRoutingTablesPerInterface creates entries in the '/etc/iproute2/rt_tables' file
for each of the interfaces passed in ifaces slice. The name of the created routing table entry will
be the same as the name of the interface. Additionally, it creates a routing rule for each of the
IP addresses assigned to each of the passed interfaces to use route table created for this specific
interface.

EnsureOnlyOneIPRuleExistsForSourceIP adds an "ip rule" entry for the SourceIP included in rule to
use the routing table RouteTableName. The routing table must have been created earlier by a call to
InitializeRoutingTablesPerInterface(). If there are already other ip rules for the same SourceIP
but different RouteTableName than in the rule, the entries are removed, so that the only currently
matching rule for SourceIP selects route table RouteTableName.

RemoveIPRuleForSourceIP removes ip rule entry for the SourceIP and RouteTableName specified in rule.
If the entry doesn't exist, no error is returned.

RemoveAllIPRulesForAddressesInSubnet removes any rules existing for any source IP within the
sourceSubnet. No error is returned if no entries were removed.

RemoveAllIPRulesForAddress removes any rules existing for the specific IP address. No error is
returned if no entries were removed.

EnsureOnlyOneIPRuleExistsForFwMark adds an "ip rule" entry for the FwMark included in rule to
use the routing table RouteTableName. The routing table must have been created earlier by a call to
InitializeRoutingTablesPerInterface(). If there are already other ip rules for the same FwMark
but different RouteTableName than in the rule, the entries are removed, so that the only currently
matching rule for FwMark selects route table RouteTableName.

EnsureRoute ensures that in the routing table tableName there is a route entry that has the form
"targetPrefix" "mode" "gateway" "options", like:
"10.10.10.0/24" "via" "10.10.20.1" ""
or
"10.10.10.0/24" "dev" "eth0" "scope link"
*/
type IPRouteHelper interface {
	InitializeRoutingTablesPerInterface(ifaces []Interface) (time.Duration, error)
	EnsureOnlyOneIPRuleExistsForSourceIP(rule IPRule) ([]IPRule, time.Duration, error)
	RemoveIPRuleForSourceIP(rule IPRule) (time.Duration, error)
	RemoveAllIPRulesForAddressesInSubnet(sourceSubnet net.IPNet) ([]IPRule, time.Duration, error)
	RemoveAllIPRulesForAddress(ip net.IP) ([]IPRule, time.Duration, error)
	EnsureOnlyOneIPRuleExistsForFwMark(rule IPRule) ([]IPRule, time.Duration, error)
	EnsureRoutes(entries []IPRouteEntry) (time.Duration, error)
}

// NewExecIPRouteHelper returns new IPRouterHelper implemented using command execution in
// the operating system.
func NewExecIPRouteHelper(executor command.Executor, ioOp SimpleFileOperator) IPRouteHelper {
	return &execIPRouteHelper{
		exec: executor,
		ioOp: ioOp,
	}
}

type execIPRouteHelper struct {
	exec command.Executor
	ioOp SimpleFileOperator
}

func (e *execIPRouteHelper) InitializeRoutingTablesPerInterface(ifaces []Interface) (time.Duration,
	error) {
	startTime := time.Now()

	for _, iface := range ifaces {
		MustExistNamedRoutingTable(iface.GetName(), e.ioOp)
		addrs, err := iface.Addrs()
		if err != nil {
			return time.Now().Sub(startTime), err
		}
		for _, addr := range addrs {
			if addr.IP.To4() == nil {
				continue
			}
			_, _, err = e.EnsureOnlyOneIPRuleExistsForSourceIP(IPRule{0, 0, addr.IP, iface.GetName()})
			if err != nil {
				return time.Now().Sub(startTime), err
			}
		}
	}
	return time.Now().Sub(startTime), nil
}

func (e *execIPRouteHelper) EnsureOnlyOneIPRuleExistsForFwMark(rule IPRule) ([]IPRule, time.Duration, error) {
	return e.ensureOnlyOneIPRuleExistsForField(rule,
		e.getAllOtherIPRulesTableNamesForFwMark,
		e.RemoveIPRuleForFwMark,
		e.ipRuleExistsForFwMark,
		[]string{"fwmark", rule.FwMarkHexed(), "lookup", rule.RouteTableName})
}

func (e *execIPRouteHelper) EnsureOnlyOneIPRuleExistsForSourceIP(rule IPRule) ([]IPRule, time.Duration, error) {
	return e.ensureOnlyOneIPRuleExistsForField(rule,
		e.getAllOtherIPRulesTableNamesForSourceIP,
		e.RemoveIPRuleForSourceIP,
		e.ipRuleExistsForSourceIP,
		[]string{"from", rule.SourceIP.String(), "lookup", rule.RouteTableName})
}

func (e *execIPRouteHelper) ensureOnlyOneIPRuleExistsForField(rule IPRule,
	getOthersFunc func(IPRule) ([]string, time.Duration, error),
	deleteRuleFunc func(IPRule) (time.Duration, error),
	checkExistsFunc func(IPRule) (bool, time.Duration, error),
	ruleAddArgs []string) ([]IPRule, time.Duration, error) {

	tableNames, tablesCheckDuration, err := getOthersFunc(rule)
	if err != nil {
		return nil, tablesCheckDuration, err
	}
	removedRules := []IPRule{}
	for _, tableName := range tableNames {
		toRemove := IPRule{0, rule.FwMark, rule.SourceIP, tableName}
		deleteRuleFunc(toRemove)
		removedRules = append(removedRules, toRemove)
	}

	// then, we make sure the necessary ip rule exists
	exists, checkDuration, checkError := checkExistsFunc(rule)
	totalDuration := tablesCheckDuration + checkDuration
	if checkError != nil {
		return removedRules, totalDuration, checkError
	}
	if exists {
		return removedRules, totalDuration, nil
	}
	log.Debugf("Adding ip rule with arguments: %s", strings.Join(ruleAddArgs, " "))
	newRuleArgs := append([]string{"rule", "add"}, ruleAddArgs...)
	res := e.exec.RunCommand("ip", newRuleArgs...)
	totalDuration += res.Duration
	if res.Err != nil {
		return nil, totalDuration, fmt.Errorf("error executing 'ip rule' command; args: %v, err: %v - %v",
			strings.Join(newRuleArgs, " "), res.Err, res.StdErr)
	}
	return removedRules, totalDuration, nil
}

func (e *execIPRouteHelper) RemoveIPRuleForSourceIP(rule IPRule) (time.Duration, error) {
	return e.removeIPRuleForField(rule, e.ipRuleExistsForSourceIP,
		[]string{"from", rule.SourceIP.String(), "lookup", rule.RouteTableName})
}

func (e *execIPRouteHelper) RemoveIPRuleForFwMark(rule IPRule) (time.Duration, error) {
	return e.removeIPRuleForField(rule, e.ipRuleExistsForFwMark,
		[]string{"fwmark", rule.FwMarkHexed(), "lookup", rule.RouteTableName})
}

func (e *execIPRouteHelper) removeIPRuleForField(rule IPRule,
	checkExistsFunc func(IPRule) (bool, time.Duration, error),
	ruleDelArgs []string) (time.Duration, error) {

	exists, checkDuration, checkError := checkExistsFunc(rule)
	if checkError != nil {
		return checkDuration, checkError
	}
	if !exists {
		return checkDuration, nil
	}
	log.Debugf("Removing ip rule with arguments: %s", strings.Join(ruleDelArgs, " "))
	delRuleArgs := append([]string{"rule", "del"}, ruleDelArgs...)
	res := e.exec.RunCommand("ip", delRuleArgs...)
	if res.Err != nil {
		return checkDuration + res.Duration, fmt.Errorf("error executing 'ip rule' command: %v - %v",
			res.Err, res.StdErr)
	}

	return checkDuration + res.Duration, nil
}

func (e *execIPRouteHelper) RemoveAllIPRulesForAddressesInSubnet(sourceSubnet net.IPNet) ([]IPRule,
	time.Duration, error) {
	return e.removeAllIPRulesMatchingPredicate(func(ruleIP net.IP, tableName string) bool {
		return sourceSubnet.Contains(ruleIP)
	})
}

func (e *execIPRouteHelper) RemoveAllIPRulesForAddress(ip net.IP) ([]IPRule, time.Duration, error) {
	return e.removeAllIPRulesMatchingPredicate(func(ruleIP net.IP, tableName string) bool {
		return ruleIP.Equal(ip)
	})
}

func (e *execIPRouteHelper) EnsureRoutes(entries []IPRouteEntry) (time.Duration, error) {
	if err := e.validateEntries(entries); err != nil {
		return 0, err
	}

	existing, duration, err := e.loadRoutingEntries(entries[0].TableName)
	if err != nil {
		return duration, err
	}

	startTime := time.Now()
	for _, entry := range entries {
		found := false
		for _, existing := range existing {
			if reflect.DeepEqual(entry, existing) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		log.Debugf("Adding routing entry: %s", entry.String())
		options := strings.Fields(entry.Options)
		args := append([]string{"route", "add", entry.TargetPrefix, entry.Mode, entry.Gateway}, options...)
		args = append(args, "tab", entry.TableName)
		res := e.exec.RunCommand("ip", args...)
		if (res.Err != nil || res.ExitCode != 0) && res.StdErr != "RTNETLINK answers: File exists\n" {
			if res.StdErr != "" {
				return duration + time.Now().Sub(startTime), fmt.Errorf("%v: %s", res.Err, res.StdErr)
			}
			return duration + time.Now().Sub(startTime), res.Err
		}
	}

	return duration + time.Now().Sub(startTime), nil
}

func (e *execIPRouteHelper) validateEntries(entries []IPRouteEntry) error {
	if len(entries) == 0 {
		return fmt.Errorf("no entries")
	}

	first := ""
	for _, e := range entries {
		if e.TableName == "" {
			return fmt.Errorf("empty string was given as routing table name")
		}
		if first == "" {
			first = e.TableName
			continue
		}
		if e.TableName != first {
			return fmt.Errorf("all the entries must be for the same routing table")
		}
	}

	return nil
}

func (e *execIPRouteHelper) loadRoutingEntries(tableName string) ([]IPRouteEntry, time.Duration, error) {
	startTime := time.Now()
	res := e.exec.RunCommand("ip", "route", "show", "tab", tableName)
	if res.Err != nil || res.ExitCode != 0 {
		return nil, res.Duration, fmt.Errorf("Error executing 'ip route show' command. ExitCode: %d, StdErr: %s",
			res.ExitCode, res.StdErr)
	}

	result := []IPRouteEntry{}
	for _, line := range strings.Split(res.StdOut, "\n") {
		if line == "" {
			continue
		}
		fields := strings.FieldsFunc(strings.Trim(line, " "), func(r rune) bool { return r == ' ' || r == '\t' })
		result = append(result, IPRouteEntry{
			TableName:    tableName,
			TargetPrefix: fields[0],
			Mode:         fields[1],
			Gateway:      fields[2],
			Options:      strings.Join(fields[3:], " "),
		})
	}

	return result, time.Now().Sub(startTime), nil
}

func (e *execIPRouteHelper) removeAllIPRulesMatchingPredicate(predicate func(ip net.IP, table string) bool) (
	[]IPRule, time.Duration, error) {
	startTime := time.Now()
	res := e.exec.RunCommand("ip", "rule")
	if res.Err != nil || res.ExitCode != 0 {
		return nil, res.Duration, fmt.Errorf("Error executing 'ip rule' command. ExitCode: %d, StdErr: %s",
			res.ExitCode, res.StdErr)
	}

	removedRules := []IPRule{}
	for _, line := range strings.Split(res.StdOut, "\n") {
		fields := strings.FieldsFunc(strings.Trim(line, " "), func(r rune) bool { return r == ' ' || r == '\t' })
		// skip empty lines
		if len(fields) != 5 {
			continue
		}
		ip := fields[2]
		if !strings.ContainsRune(ip, '/') {
			ip += "/32"
		}
		ruleSrcIP, _, err := net.ParseCIDR(ip)
		if err != nil {
			continue
		}
		ruleTable := fields[4]
		if predicate(ruleSrcIP, ruleTable) {
			log.Debugf("Removing ip route rule: %s", line)
			cmdArgs := append([]string{"rule", "del"}, fields[1:]...)
			res = e.exec.RunCommand("ip", cmdArgs...)
			removedRules = append(removedRules, IPRule{0, 0, ruleSrcIP, fields[4]})
		}
	}
	return removedRules, time.Now().Sub(startTime), nil
}

func (e *execIPRouteHelper) execIPRulePipe(pipeExpression string) (*command.ExecResult, error) {
	checkRule := fmt.Sprintf("ip rule | %s", pipeExpression)
	res := e.exec.RunCommand("sh", "-c", checkRule)
	if res.ExitCode != 0 && res.ExitCode != 1 && res.Err != nil {
		return res, fmt.Errorf("error executing 'ip rule' command; err: %v; stdErr: %v, stdOut: %v; exitCode: %d",
			res.Err, res.StdErr, res.StdOut, res.ExitCode)
	}

	return res, nil
}

func (e *execIPRouteHelper) ipRuleExistsForSourceIP(rule IPRule) (bool, time.Duration, error) {
	grepSourceIP, grepTableName := rule.GrepEscaped()

	// check if the rule already exists with
	// sh -c 'ip ru | grep "from 10.69.10.210 lookup 10.69.10.211 $"'
	args := fmt.Sprintf("grep \"from %s lookup %s $\"", grepSourceIP, grepTableName)
	res, err := e.execIPRulePipe(args)
	if err != nil {
		return false, res.Duration, err
	}
	return res.ExitCode == 0, res.Duration, nil
}

func (e *execIPRouteHelper) ipRuleExistsForFwMark(rule IPRule) (bool, time.Duration, error) {
	// check if the rule already exists with
	// sh -c 'ip ru | grep "fwmark 0x100 lookup 100 $"'
	args := fmt.Sprintf("grep \"fwmark %s lookup %s $\"", rule.FwMarkHexed(), rule.RouteTableName)
	res, err := e.execIPRulePipe(args)
	if err != nil {
		return false, res.Duration, err
	}
	return res.ExitCode == 0, res.Duration, nil
}

func (e *execIPRouteHelper) getAllOtherIPRulesTableNamesForSourceIP(rule IPRule) ([]string,
	time.Duration, error) {
	// ip rule | grep -v "from 10\.111\.133\.165 lookup 10\.69\.142\.48" | grep "from 10\.111\.133\.165 lookup "
	// output:
	// 27320:	from 10.69.128.18 lookup 10.69.128.18
	grepSourceIP, grepTableName := rule.GrepEscaped()
	args := fmt.Sprintf("grep -v \"from %s lookup %s $\" | grep \"from %s lookup \"",
		grepSourceIP, grepTableName, grepSourceIP)
	res, err := e.execIPRulePipe(args)
	if err != nil {
		return nil, res.Duration, err
	}
	result := []string{}
	for _, line := range strings.Split(res.StdOut, "\n") {
		out := strings.Fields(line)
		if len(out) == 5 {
			result = append(result, out[4])
		}
	}
	return result, res.Duration, nil
}

func (e *execIPRouteHelper) getAllOtherIPRulesTableNamesForFwMark(rule IPRule) ([]string,
	time.Duration, error) {
	// ip rule | grep -v "fwmark 0x100 lookup 100" | grep "fwmark 0x100 lookup "
	// output:
	// 32761:	from 127.0.0.1 fwmark 0x100 lookup 101
	args := fmt.Sprintf("grep -v \"fwmark %s lookup %s $\" | grep \"fwmark %s lookup \"",
		rule.FwMarkHexed(), rule.RouteTableName, rule.FwMarkHexed())
	res, err := e.execIPRulePipe(args)
	if err != nil {
		return nil, res.Duration, err
	}
	result := []string{}
	for _, line := range strings.Split(res.StdOut, "\n") {
		out := strings.Fields(line)
		if len(out) == 7 {
			result = append(result, out[6])
		}
	}
	return result, res.Duration, nil
}
