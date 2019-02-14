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
	"github.com/DevFactory/go-tools/pkg/linux/command"
)

func ExecResultOKNoOutput() *command.ExecResult {
	return ExecResultExitCodeNoOutput(0)
}

func ExecResultExitCodeNoOutput(exitCode int) *command.ExecResult {
	return ExecResultExitCodeStdOutput(exitCode, "")
}

func ExecResultExitCodeStdOutput(exitCode int, stdOut string) *command.ExecResult {
	return ExecResultExitCodeStdErrOutput(exitCode, stdOut, "")
}

func ExecResultExitCodeErrOutput(exitCode int, errOut string) *command.ExecResult {
	return ExecResultExitCodeStdErrOutput(exitCode, "", errOut)
}

func ExecResultExitCodeStdErrOutput(exitCode int, stdOut, errOut string) *command.ExecResult {
	return &command.ExecResult{Duration: 1, Err: nil, ExitCode: exitCode, StdOut: stdOut, StdErr: errOut}
}

func ExecResultGrepNotFound() *command.ExecResult {
	return ExecResultExitCodeNoOutput(1)
}

func ExecResultGrepFound(stdOut string) *command.ExecResult {
	return ExecResultExitCodeStdOutput(0, stdOut)
}
