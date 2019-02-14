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
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/DevFactory/go-tools/pkg/linux/command"
)

/*
ConntrackHelper provides operations backed up by the system "conntrack" command

RemoveEntriesDNATingToIP removes all entries in the conntrack table, that describe connections running with
DNAT and having destination IP equal to ip. If returned error is nil, then time.Duration returned type
is used to let the caller know how long it took to execute the command.
*/
type ConntrackHelper interface {
	RemoveEntriesDNATingToIP(ip net.IP) (time.Duration, error)
}

// NewExecConntrackHelper creates new ConntrackHelper implemented by executing linux
// `conntrack` command
func NewExecConntrackHelper(executor command.Executor) ConntrackHelper {
	return &execConntrackHelper{
		exec: executor,
	}
}

type execConntrackHelper struct {
	exec command.Executor
}

func (c *execConntrackHelper) RemoveEntriesDNATingToIP(ip net.IP) (time.Duration, error) {
	log.Debugf("Removing all conntrack entries DNATing to %s", ip)

	args := []string{"-D", "-g", "-r", ip.String()}
	res := c.exec.RunCommand("conntrack", args...)

	if res.Err != nil && strings.Index(res.StdErr, "0 flow entries have been deleted") == -1 {
		return res.Duration, fmt.Errorf("error executing 'conntrack' command; args: %v, err: %v - %v",
			strings.Join(args, " "), res.Err, res.StdErr)
	}

	return res.Duration, nil
}
