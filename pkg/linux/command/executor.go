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

package command

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// ExecResult returns the outcome of linux command execution
type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
	Duration time.Duration
	Err      error
}

// Executor allows you to run operating system commands from go 
type Executor interface {
	// RunCommand runs a command once and returns ExecResult
	RunCommand(command string, args ...string) *ExecResult

	// RunCommandWithRetries runs a command until the ExecResult.ExitCode is in okExitCodes or 
	// the maximum number of retries is reached. Commands are executed one after another without 
	// any delay.
	RunCommandWithRetries(retries int, okExitCodes []int, command string, args ...string) *ExecResult

	// RunCommandWithRetriesAndDelay runs a command until the ExecResult.ExitCode is in okExitCodes or 
	// the maximum number of retries is reached. Commands are executed one after another with a delay
	// of retryWaitMilliseconds between each execution.
	RunCommandWithRetriesAndDelay(retries int, retryWaitMilliseconds int, okExitCodes []int,
		command string, args ...string) *ExecResult
}

// NewExecutor returns a new Linux command executor.
func NewExecutor() Executor {
	return &LinuxExecutor{}
}

// LinuxExecutor implements Executor for Linux
type LinuxExecutor struct {
}

// RunCommand runs a command once and returns ExecResult
func (l *LinuxExecutor) RunCommand(command string, args ...string) *ExecResult {
	cmd := exec.Command(command, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	timeStart := time.Now()
	if err := cmd.Start(); err != nil {
		return &ExecResult{"", "", -1, 0, fmt.Errorf("couldn't start a command %v %v: %v", command, args, err)}
	}

	err := cmd.Wait()
	duration := time.Now().Sub(timeStart)
	if err != nil {
		var exiterr *exec.ExitError
		var ok bool
		if exiterr, ok = err.(*exec.ExitError); ok {
			// The program has successfully ran and exited with an exit code != 0
			status, _ := exiterr.Sys().(syscall.WaitStatus)
			return &ExecResult{outBuf.String(), errBuf.String(), status.ExitStatus(), duration, exiterr}
		}
		// the program failed for other reason than signalled with exit code
		return &ExecResult{outBuf.String(), errBuf.String(), -1, duration,
			fmt.Errorf("unexpected error waiting for exit of command %v %v: %v", command, args, exiterr)}
	}
	return &ExecResult{outBuf.String(), errBuf.String(), 0, duration, nil}
}

// RunCommandWithRetriesAndDelay runs a command until the ExecResult.ExitCode is in okExitCodes or 
// the maximum number of retries is reached. Commands are executed one after another with a delay
// of retryWaitMilliseconds between each execution.
func (l *LinuxExecutor) RunCommandWithRetriesAndDelay(retries int, retryWaitMilliseconds int,
	okExitCodes []int, command string, args ...string) *ExecResult {
	var result *ExecResult
	for i := 0; i < retries; i++ {
		result = l.RunCommand(command, args...)
		if result.Err == nil {
			return result
		}
		for _, ec := range okExitCodes {
			if ec == result.ExitCode {
				result.Err = nil
				return result
			}
		}
		time.Sleep(time.Duration(retryWaitMilliseconds) * time.Millisecond)
	}
	// if retries were unsuccessful, return last error
	return result
}

// RunCommandWithRetries runs a command until the ExecResult.ExitCode is in okExitCodes or 
// the maximum number of retries is reached. Commands are executed one after another without 
// any delay.
func (l *LinuxExecutor) RunCommandWithRetries(retries int, okExitCodes []int, command string,
	args ...string) *ExecResult {
	return l.RunCommandWithRetriesAndDelay(retries, 0, okExitCodes, command, args...)
}
