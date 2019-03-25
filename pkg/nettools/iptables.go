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
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/DevFactory/go-tools/pkg/linux/command"
)

var (
	commentTemplate   = "-m comment --comment \"(%s)\" "
	commentArgsRegexp = regexp.MustCompile(fmt.Sprintf(commentTemplate, ".*"))
)

const (
	iptablesRetries          = 3
	iptablesRetriesDelayMSec = 100
	// this uses awk to list the content of a single chain in a table using iptables-save command
	awkIptablesSaveMagicFilter = "iptables-save | awk -v table=%s -v chain=%s " +
		`'$0 ~ "^\*"table"$" {in_table=1};$1 ~ "^COMMIT$" {in_table=0};in_table == 1 && $2 ~ "^"chain"$" {print $0}'`
)

// IPTablesRuleArgs provides arguments for an iptables rule
type IPTablesRuleArgs struct {
	Table     string
	ChainName string
	Selector  []string
	Action    []string
	Comment   string
}

// GetSelectorAndAction returns strings that are ready to be used with
// iptables command for selector and action
func (args IPTablesRuleArgs) GetSelectorAndAction() (string, string) {
	selector := strings.Join(args.Selector, " ")
	action := strings.Join(args.Action, " ")

	return selector, action
}

// IPTablesHelper provides interface to some operations on system's firewall system using the
// iptables command
type IPTablesHelper interface {
	// EnsureChainExists makes sure the chain customChainName exists in table
	EnsureChainExists(table, customChainName string) error
	// EnsureJumpToChainExists makes sure a jump instruction from baseChainName to
	// customChainName exists in table
	EnsureJumpToChainExists(table, customChainName, baseChainName string) error
	// EnsureExistsInsert makes sure the rule exists. It is inserted at the start of
	// the chain, if the exact same rule doesn't exist; nothing happens otherwise.
	EnsureExistsInsert(args IPTablesRuleArgs) error
	// EnsureExistsAppend makes sure the rule exists. It is appended at the end of
	// the chain, if the exact same rule doesn't exist; nothing happens otherwise.
	EnsureExistsAppend(args IPTablesRuleArgs) error
	// EnsureExistsOnlyAppend checks if the exact same rule already exists. If not,
	// the rule is appended to the selected chain. If other rules with exactly the
	// same comment exist, they are removed.
	EnsureExistsOnlyAppend(args IPTablesRuleArgs) error
	// Delete removes the exact same rule. Returns error if it doesn't exist.
	Delete(args IPTablesRuleArgs) error
	// DeleteByComment removes all rules in the specified table and chain that have
	// the same comment.
	DeleteByComment(table, chain, comment string) error
	// FlushChain removes all rules in chain chainName in table tableName. Returns error
	// if the chain doesn't exist.
	FlushChain(tableName, chainName string) error
	// DeleteChain removes the chain chainName in table tableName. Returns error
	// if the chain doesn't exist.
	DeleteChain(tableName, chainName string) error
	// LoadRules loads all rules from the specified table and chain. Returns error
	// if the chain doesn't exist.
	LoadRules(tableName, chainName string) ([]*IPTablesRuleArgs, error)
}

type execIPTablesHelper struct {
	exec     command.Executor
	lockTime time.Duration
}

// NewExecIPTablesHelper returns IPTablesHelper implemented by executing linux iptables
// command
func NewExecIPTablesHelper(exec command.Executor, lockTime time.Duration) IPTablesHelper {
	return &execIPTablesHelper{
		exec:     exec,
		lockTime: lockTime,
	}
}

func (h *execIPTablesHelper) EnsureChainExists(table, customChainName string) error {
	msg := "chain " + customChainName
	return h.checkAndExecRule(table, customChainName, "", "", "", msg, "creating", true,
		func() error {
			return h.runChangingRule(table, customChainName, "-N", "", "", "", nil)
		})
}

func (h *execIPTablesHelper) EnsureJumpToChainExists(table, customChainName, baseChainName string) error {
	msg := fmt.Sprintf("jump from chain %s to chain %s", baseChainName, customChainName)
	return h.checkAndExecRule(table, baseChainName, "", "for SNM", customChainName, msg, "creating", true,
		func() error {
			return h.runChangingRule(table, baseChainName, "-A", "", "for SNM", customChainName, nil)
		})
}

func (h *execIPTablesHelper) EnsureExistsInsert(args IPTablesRuleArgs) error {
	return h.ensureExistsWithOption(args, "-I")
}

func (h *execIPTablesHelper) EnsureExistsAppend(args IPTablesRuleArgs) error {
	return h.ensureExistsWithOption(args, "-A")
}

func (h *execIPTablesHelper) EnsureExistsOnlyAppend(args IPTablesRuleArgs) error {
	rules, err := h.loadRulesWithComment(args.Table, args.ChainName, args.Comment)
	if err != nil {
		return err
	}
	ruleExists := false
	for _, rule := range rules {
		// if we found exactly the same rule, set ruleExists and go checking other rules
		if reflect.DeepEqual(args, *rule) {
			ruleExists = true
			continue
		}
		selector, action := rule.GetSelectorAndAction()
		err = h.runChangingRule(rule.Table, rule.ChainName, "-D", selector, rule.Comment, action, nil)
		if err != nil {
			log.Debug("Error deleting rule by comment in table %s chain %s; exact info above; error: %v",
				args.Table, args.ChainName, "-D", selector, action)
			return err
		}
	}
	if ruleExists {
		return nil
	}
	selector, action := args.GetSelectorAndAction()
	err = h.runChangingRule(args.Table, args.ChainName, "-A", selector, args.Comment, action, nil)
	return err
}

func (h *execIPTablesHelper) Delete(args IPTablesRuleArgs) error {
	subj := fmt.Sprintf("rule in chain %s with selector %s and action %s",
		args.ChainName, args.Selector, args.Action)
	selector, action := args.GetSelectorAndAction()
	return h.checkAndExecRule(args.Table, args.ChainName, selector, args.Comment, action, subj, "deleting", false,
		func() error {
			return h.runChangingRule(args.Table, args.ChainName, "-D", selector, args.Comment, action, nil)
		})
}

func (h *execIPTablesHelper) DeleteByComment(table, chain, comment string) error {
	rules, err := h.loadRulesWithComment(table, chain, comment)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		selector, action := rule.GetSelectorAndAction()
		err = h.runChangingRule(rule.Table, rule.ChainName, "-D", selector, rule.Comment, action, nil)
		if err != nil {
			log.Debug("Error deleting rule by comment in table %s chain %s; exact info above; error: %v",
				table, chain, err)
			return err
		}
	}
	return nil
}

func (h *execIPTablesHelper) FlushChain(tableName, chainName string) error {
	return h.runChangingRule(tableName, chainName, "-F", "", "", "", []int{0, 1})
}

func (h *execIPTablesHelper) DeleteChain(tableName, chainName string) error {
	return h.runChangingRule(tableName, chainName, "-X", "", "", "", []int{0, 1})
}

func (h *execIPTablesHelper) LoadRules(tableName, chainName string) ([]*IPTablesRuleArgs, error) {
	return h.loadRulesWithComment(tableName, chainName, "")
}

func (h *execIPTablesHelper) checkAndExecRule(tableName, chainName, selector, comment,
	action, debugSubject, debugAction string, shouldExist bool, actionFunc func() error) error {
	exists, res := h.runExistsRule(tableName, chainName, selector, comment, action)
	if res != nil && res.Err != nil {
		log.Debugf("Error checking for %s in table %s: exitCode: %d, stdErr: %s, err: %v",
			debugSubject, tableName, res.ExitCode, res.StdErr, res.Err)
		return res.Err
	}
	if (exists && shouldExist) || (!exists && !shouldExist) {
		return nil
	}
	err := actionFunc()
	if err != nil {
		log.Debugf("Error %s %s in table %s: %v", debugAction, debugSubject, tableName, err)
	}
	return err
}

func (h *execIPTablesHelper) runChangingRule(tableName, chainName, option, selector, comment,
	action string, okExitCodes []int) error {
	log.Debugf("Executing iptables in table %s and chain %s; rule option %s, selector: %s, action: %s, "+
		"comment: %s", tableName, chainName, option, selector, action, comment)
	res := h.runRule(tableName, chainName, option, selector, comment, action, okExitCodes)
	if res.Err != nil || res.ExitCode != 0 {
		log.Debugf("iptables execution failed in table %s and chain %s; rule selector: %s, action: %s;"+
			" exitCode: %d, stdErr: %s, err: %v", tableName, chainName, selector, action, res.ExitCode,
			res.StdErr, res.Err)
		if res.Err != nil {
			return res.Err
		}
		errBuf := make([]byte, len(res.StdErr))
		copy(errBuf, res.StdErr)
		return &exec.ExitError{
			Stderr: errBuf,
		}
	}
	return nil
}

func (h *execIPTablesHelper) ensureExistsWithOption(args IPTablesRuleArgs, option string) error {
	msg := fmt.Sprintf("rule in chain %s with selector %s and action %s",
		args.ChainName, args.Selector, args.Action)
	selector, action := args.GetSelectorAndAction()
	return h.checkAndExecRule(args.Table, args.ChainName, selector, args.Comment, action, msg, "creating", true,
		func() error {
			return h.runChangingRule(args.Table, args.ChainName, option, selector, args.Comment, action, nil)
		})
}

func (h *execIPTablesHelper) runRule(tableName, chainName, option, selector, comment,
	action string, okExitCodes []int) *command.ExecResult {
	cmd := h.getRuleCommand(tableName, chainName, option, selector, comment, action)
	res := h.exec.RunCommandWithRetriesAndDelay(iptablesRetries, iptablesRetriesDelayMSec, okExitCodes,
		cmd[0], cmd[1:]...)
	return res
}

func (h *execIPTablesHelper) getRuleCommand(tableName, chainName, option, selector, comment, action string) []string {
	command := []string{"iptables"}
	if option == "-L" {
		command = append(command, "-n")
	}
	command = append(command, "-w", fmt.Sprintf("%d", int(h.lockTime.Seconds())), "-t", tableName, option)
	if chainName != "" {
		command = append(command, chainName)
	}
	if selector != "" {
		command = append(command, strings.Split(selector, " ")...)
	}
	if comment != "" {
		command = append(command, "-m", "comment", "--comment", fmt.Sprintf("\"%s\"", comment))
	}
	if action != "" {
		command = append(command, "-j")
		command = append(command, strings.Split(action, " ")...)
	}

	return command
}

func (h *execIPTablesHelper) runExistsRule(tableName, chainName, selector, comment, action string) (bool,
	*command.ExecResult) {
	option := "-C"
	if selector == "" && action == "" {
		option = "-L"
	}
	cmd := h.getRuleCommand(tableName, chainName, option, selector, comment, action)
	res := h.exec.RunCommandWithRetriesAndDelay(iptablesRetries, iptablesRetriesDelayMSec, []int{0, 1},
		cmd[0], cmd[1:]...)
	if res.ExitCode == 1 {
		return false, nil
	}
	if res.ExitCode == 2 && strings.Contains(res.StdErr, " doesn't exist.\n") {
		return false, nil
	}
	if res.ExitCode == 0 {
		return true, nil
	}
	return false, res
}

func (h *execIPTablesHelper) listRules(tableName, chainName, regexpFilter string) ([]string, error) {
	shCommand := fmt.Sprintf(awkIptablesSaveMagicFilter, tableName, chainName)
	res := h.exec.RunCommandWithRetriesAndDelay(iptablesRetries, iptablesRetriesDelayMSec, []int{0},
		"sh", "-c", shCommand)
	if res.Err != nil || res.StdErr != "" {
		log.Errorf("Error running iptables-save with awk filter for table %s and chain %s: %v - %v",
			tableName, chainName, res.Err, res.StdErr)
		if res.Err != nil {
			return nil, res.Err
		}
		errBuf := make([]byte, len(res.StdErr))
		copy(errBuf, res.StdErr)
		return nil, &exec.ExitError{
			Stderr: errBuf,
		}
	}
	// iptables-save wraps comment string into its own pair of "" and additionally escapes ones
	// form the comment into \"
	output := strings.Replace(res.StdOut, `\"`, "", -1)
	entries := strings.FieldsFunc(output, func(c rune) bool {
		return c == '\n'
	})
	if regexpFilter == "" {
		return entries, nil
	}
	result := []string{}
	regexp := regexp.MustCompile(regexpFilter)
	for _, entry := range entries {
		if regexp.MatchString(entry) {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (h *execIPTablesHelper) parseIPTablesSaveEntry(table, chain, entry string) (*IPTablesRuleArgs, error) {
	// the entry must be created with an iptables command of the form:
	// iptables -n -t [table] -[option] [selector] (-m comment --comment [cmt]) -j [action]
	// examples:
	// -A test -s 127.0.0.1/32 -p tcp -m tcp --dport 1717 -m comment --comment "this is a comment 3" -j DNAT --to-destination 127.0.0.2:5454
	// -A MAP-NWKDVBYOKRF4ZB7BZZV2RAJU -m comment --comment "\"mark for masquerade in SNM-POSTROUTING-MASQ\"" -j MARK --set-xmark 0x100000/0x100000
	result := &IPTablesRuleArgs{
		Table:     table,
		ChainName: chain,
	}
	indexPairs := commentArgsRegexp.FindStringSubmatchIndex(entry)
	if indexPairs != nil {
		if len(indexPairs) != 4 {
			return nil, fmt.Errorf("can't parse entry as can't correctly identify comment section")
		}
		// it has comment - get its value
		result.Comment = strings.TrimSpace(entry[indexPairs[2]:indexPairs[3]])
		// and remove from the entry
		entry = entry[:indexPairs[0]] + entry[indexPairs[1]:]
	}
	// action is everything after " -j "
	index := strings.Index(entry, " -j ")
	if index == -1 {
		return nil, fmt.Errorf("can't parse entry as can't find action section")
	}
	result.Action = strings.Split(entry[index+4:], " ")
	selector := ""
	selectorStart := 4 + len(chain)
	if index > selectorStart {
		selector = entry[selectorStart:index]
	}
	// iptables injects pointless "-m tcp" or "-m udp"
	selector = strings.Replace(selector, " -m tcp", "", 1)
	selector = strings.Replace(selector, " -m udp", "", 1)
	result.Selector = strings.Fields(selector)

	return result, nil
}

func (h *execIPTablesHelper) loadRulesWithComment(tableName, chainName, comment string) (
	[]*IPTablesRuleArgs, error) {
	filter := ""
	if comment != "" {
		filter = fmt.Sprintf(commentTemplate, regexp.QuoteMeta(comment))
	}
	entries, err := h.listRules(tableName, chainName, filter)
	if err != nil {
		return nil, err
	}
	result := make([]*IPTablesRuleArgs, len(entries))
	for i, entry := range entries {
		rule, err := h.parseIPTablesSaveEntry(tableName, chainName, entry)
		if err != nil {
			log.Debug("Can't parse rules loaded from table %s and chain %s", tableName, chainName)
		}
		result[i] = rule
	}
	return result, nil
}
