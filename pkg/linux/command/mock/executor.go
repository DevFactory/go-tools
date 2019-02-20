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

package mock

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DevFactory/go-tools/pkg/linux/command"
	"github.com/stretchr/testify/assert"
)

type ExecInfo struct {
	Expected string
	Returned *command.ExecResult
}

type ExecInfoList []ExecInfo

type ExecProviderFunction func(int, *testing.T, string, ...string) *command.ExecResult

type MockExecutor struct {
	CallNum            int
	t                  *testing.T
	info               []*ExecInfo
	unexpectedCommands []string
}

func NewMockExecutorFromInfos(t *testing.T, execInfos ...*ExecInfo) *MockExecutor {
	info := make([]*ExecInfo, 0)
	info = append(info, execInfos...)
	return &MockExecutor{
		CallNum:            0,
		info:               info,
		t:                  t,
		unexpectedCommands: make([]string, 0),
	}
}

func NewMockExecutorFromSlices(t *testing.T, expected []string, returned []*command.ExecResult) (*MockExecutor, error) {
	if len(expected) != len(returned) {
		return nil, fmt.Errorf("expected and returned lists have different lengths")
	}
	info := make([]*ExecInfo, 0, len(expected))
	for i := range expected {
		info = append(info, &ExecInfo{expected[i], returned[i]})
	}
	return &MockExecutor{
		CallNum:            0,
		info:               info,
		t:                  t,
		unexpectedCommands: make([]string, 0),
	}, nil
}

func (e *MockExecutor) RunCommand(cmd string, args ...string) *command.ExecResult {
	var ret *command.ExecResult
	if e.CallNum < len(e.info) {
		e.validateExecCall(e.t, e.info[e.CallNum].Expected, cmd, args...)
		ret = e.info[e.CallNum].Returned
	} else {
		e.unexpectedCommands = append(e.unexpectedCommands, cmd+" "+strings.Join(args, " "))
		ret = e.cmdPrinter(e.CallNum, e.t, cmd, args...)
	}
	e.CallNum++
	return ret
}

func (e *MockExecutor) RunCommandWithRetries(retries int, okExitCodes []int, cmd string,
	args ...string) *command.ExecResult {
	return e.RunCommandWithRetriesAndDelay(retries, 0, okExitCodes, cmd, args...)
}

func (e *MockExecutor) RunCommandWithRetriesAndDelay(retries int, retryWaitMilliseconds int,
	okExitCodes []int, cmd string, args ...string) *command.ExecResult {
	var ret *command.ExecResult
	for i := 0; i < retries; i++ {
		ret = e.RunCommand(cmd, args...)
		if ret.Err == nil {
			return ret
		}
		for _, ec := range okExitCodes {
			if ec == ret.ExitCode {
				ret.Err = nil
				return ret
			}
		}
		time.Sleep(time.Duration(retryWaitMilliseconds) * time.Second / 1e3)
	}
	return ret
}

func (e *MockExecutor) ValidateCallNum() {
	msg := ""
	if e.CallNum < len(e.info) {
		msg = fmt.Sprintf("%d expected commands were never called:\n", len(e.info)-e.CallNum)
		for _, cmd := range e.info[e.CallNum:] {
			msg = msg + cmd.Expected + "\n"
		}
	}
	if e.CallNum > len(e.info) {
		msg = fmt.Sprintf("%d unexpected commands were called:\n", len(e.unexpectedCommands))
		for _, cmd := range e.unexpectedCommands {
			msg = msg + cmd + "\n"
		}
	}
	assert.Equal(e.t, len(e.info), e.CallNum, msg)
}

func (e *MockExecutor) validateExecCall(t *testing.T, expectedCommand, realCommand string, realArgs ...string) {
	rArgs := make([]string, 0, len(realArgs))
	for i := range realArgs {
		rArgs = append(rArgs, strings.Split(realArgs[i], " ")...)
	}
	splitCmd := strings.Split(expectedCommand, " ")

	if !assert.Equal(t, len(splitCmd)-1, len(rArgs)) {
		actualCommand := fmt.Sprintf("%s %s", realCommand, strings.Join(rArgs, " "))
		t.Fatalf("Can't compare real command to expected one, as they have different "+
			"number of arguments. Expected command:\n%s\nActual command:\n%s\n",
			expectedCommand, actualCommand)
		t.Fail()
		return
	}
	assert.Equal(t, splitCmd[0], realCommand)
	for i := range rArgs {
		assert.Equal(t, splitCmd[i+1], rArgs[i])
	}
}

func (e *MockExecutor) cmdPrinter(callNum int, t *testing.T, cmd string, args ...string) *command.ExecResult {
	fmt.Printf("%s %s\n", cmd, strings.Join(args, " "))
	return &command.ExecResult{
		StdOut:   "",
		StdErr:   "",
		ExitCode: 1,
		Duration: time.Nanosecond,
		Err:      nil,
	}
}
