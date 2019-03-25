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
	"os/exec"
	"strings"

	"github.com/DevFactory/go-tools/pkg/extensions/collections"
	"github.com/DevFactory/go-tools/pkg/linux/command"
	log "github.com/sirupsen/logrus"
)

const (
	ipsetListWithAwk = "ipset list %s | awk " + `'$0 ~ "^Members:$" {found=1; ln=NR}; NR>ln && found == 1 {print $1}'`
)

/*
IPSetHelper provides methods to manage ipset sets.

EnsureSetExists creates the named ipset if it doesn't exist.

DeleteSet removes the ipset with name. Returns error if it doesn't exist.

EnsureSetHasOnly makes sure that IP addresses from ips are the only ones present in the set
name.

GetIPs returns a slice of IP addresses from set name. Returns error if it doesn't exist.
*/
type IPSetHelper interface {
	EnsureSetExists(name, setType string) error
	DeleteSet(name string) error
	EnsureSetHasOnly(name string, ips []net.IP) error
	GetIPs(name string) ([]net.IP, error)
}

type execIPSetHelper struct {
	exec command.Executor
}

// NewExecIPSetHelper returns new IPSetHelper based on Linux command `ipset`
func NewExecIPSetHelper(exec command.Executor) IPSetHelper {
	return &execIPSetHelper{
		exec: exec,
	}
}

func (h *execIPSetHelper) EnsureSetExists(name, setType string) error {
	// ipset create [name] [type=hash:ip] comment counters -exist
	log.Debugf("Ensuring ipset %s exists", name)
	res := h.exec.RunCommand("ipset", "create", name, setType, "comment", "counters", "-exist")
	if res.Err != nil || res.ExitCode != 0 {
		log.Debugf("Error creating/checking ipset %s of type %s: %v, stdErr: %s",
			name, setType, res.Err, res.StdErr)
		return res.Err
	}
	return nil
}

func (h *execIPSetHelper) DeleteSet(name string) error {
	// ipset destroy [name]
	log.Debugf("Destroying ipset %s", name)
	res := h.exec.RunCommand("ipset", "destroy", name)
	if res.Err != nil || res.ExitCode != 0 {
		log.Debugf("Error removing ipset %s: %v, stdErr: %s",
			name, res.Err, res.StdErr)
		if res.StdErr != "" {
			if ee, ok := res.Err.(*exec.ExitError); ok {
				errBuf := make([]byte, len(res.StdErr))
				copy(errBuf, res.StdErr)
				ee.Stderr = errBuf
			}
		}
		return res.Err
	}
	return nil
}

func (h *execIPSetHelper) EnsureSetHasOnly(name string, ips []net.IP) error {
	// load the current set and assume tentatively all IPs from it are going
	// to be removed
	current, err := h.GetIPs(name)
	if err != nil {
		return err
	}
	newAsInterface := make([]interface{}, len(ips))
	for i, ip := range ips {
		newAsInterface[i] = ip
	}
	currentAsInterface := make([]interface{}, len(current))
	for i, ip := range current {
		currentAsInterface[i] = ip
	}
	// find the diff
	toAdd, toRemove := collections.GetSlicesDifferences(newAsInterface, currentAsInterface,
		func(ip1, ip2 interface{}) bool {
			return (ip1.(net.IP)).Equal(ip2.(net.IP))
		})

	for _, iip := range toAdd {
		ip := iip.(net.IP)
		log.Debugf("Adding IP %s to ipset %s", ip.String(), name)
		if err := h.addIPToSet(name, ip); err != nil {
			log.Errorf("Error adding entry %v to ipset %s", ip, name)
			return err
		}
	}
	for _, iip := range toRemove {
		ip := iip.(net.IP)
		log.Debugf("Removing IP %s from ipset %s", ip.String(), name)
		if err := h.removeIPFromSet(name, ip); err != nil {
			log.Debugf("Error removing entry %v from ipset %s", ip, name)
			return err
		}
	}

	return nil
}

func (h *execIPSetHelper) GetIPs(name string) ([]net.IP, error) {
	// # ipset list myset | awk '$0 ~ "^Members:$" {found=1; ln=NR}; NR>ln && found == 1 {print $1}'
	// 127.0.0.1
	// 127.0.0.2
	cmd := fmt.Sprintf(ipsetListWithAwk, name)
	res := h.exec.RunCommand("sh", "-c", cmd)
	if res.Err != nil || res.ExitCode != 0 {
		log.Debugf("Problem listing ipset %s - probably it's OK and it just doesn't exist: "+
			"%v, stdErr: %s", name, res.Err, res.StdErr)
		return []net.IP{}, res.Err
	}
	lines := strings.Split(res.StdOut, "\n")
	result := make([]net.IP, 0, len(lines))
	for _, line := range lines {
		ip := net.ParseIP(strings.TrimSpace(line))
		if ip != nil {
			result = append(result, ip)
		}
	}
	return result, nil
}

func (h *execIPSetHelper) addIPToSet(name string, ip net.IP) error {
	res := h.exec.RunCommand("ipset", "add", name, ip.String())
	if res.Err != nil || res.ExitCode != 0 {
		log.Errorf("Error adding IP %s to ipset %s: %v, stdErr: %s",
			ip.String(), name, res.Err, res.StdErr)
		return res.Err
	}
	return nil
}

func (h *execIPSetHelper) removeIPFromSet(name string, ip net.IP) error {
	res := h.exec.RunCommand("ipset", "del", name, ip.String())
	if res.Err != nil || res.ExitCode != 0 {
		log.Debugf("Error removing IP %s from ipset %s: %v, stdErr: %s",
			ip.String(), name, res.Err, res.StdErr)
		return res.Err
	}
	return nil
}
