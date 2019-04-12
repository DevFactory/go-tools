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
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/DevFactory/go-tools/pkg/extensions/collections"
	"github.com/DevFactory/go-tools/pkg/linux/command"
	log "github.com/sirupsen/logrus"
)

const (
	// IPSetListWithAwk is a string to execute an ipset list command and filter out results with awk
	IPSetListWithAwk = "ipset list %s | awk " + `'$0 ~ "^Members:$" {found=1; ln=NR}; NR>ln && found == 1 {print $1}'`
)

/*
NetPort allows to store net.IPNet and a port number with protocol for ipsets based on that data.
*/
type NetPort struct {
	Net      net.IPNet
	Protocol Protocol
	Port     uint16
}

// String returns a format accepted by ipset, ie.
// 10.0.0.0/8,tcp:80
func (np NetPort) String() string {
	return fmt.Sprintf("%s,%s:%d", np.Net.String(), np.Protocol, np.Port)
}

// Equal returns true only if the NetPort has exactly the same values as the parameter NetPort.
func (np NetPort) Equal(np2 NetPort) bool {
	if np.Net.IP.Equal(np2.Net.IP) && bytes.Compare(np.Net.Mask, np2.Net.Mask) == 0 &&
		np.Port == np2.Port && np.Protocol == np2.Protocol {
		return true
	}
	return false
}

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
	EnsureSetHasOnlyNetPort(name string, netports []NetPort) error
	GetNetPorts(name string) ([]NetPort, error)
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
	return h.ensureSetHasOnlyGeneric(name, "IP", ipSliceToInterface(ips),
		func(setName string) ([]interface{}, error) {
			ips, err := h.GetIPs(setName)
			return ipSliceToInterface(ips), err
		},
		func(e1, e2 interface{}) bool {
			return (e1.(net.IP)).Equal(e2.(net.IP))
		},
		func(setName string, obj interface{}) error {
			return h.addElementToSet(name, "IP", obj.(net.IP))
		},
		func(setName string, obj interface{}) error {
			return h.removeElementFromSet(name, "IP", obj.(net.IP))
		})
}

func (h *execIPSetHelper) EnsureSetHasOnlyNetPort(name string, netports []NetPort) error {
	return h.ensureSetHasOnlyGeneric(name, "NetPort", netPortSliceToInterface(netports),
		func(setName string) ([]interface{}, error) {
			nps, err := h.GetNetPorts(setName)
			return netPortSliceToInterface(nps), err
		},
		func(e1, e2 interface{}) bool {
			return e1.(NetPort).Equal(e2.(NetPort))
		},
		func(setName string, obj interface{}) error {
			return h.addElementToSet(name, "NetPort", obj.(NetPort))
		},
		func(setName string, obj interface{}) error {
			return h.removeElementFromSet(name, "NetPort", obj.(NetPort))
		})
}

func (h *execIPSetHelper) GetIPs(name string) ([]net.IP, error) {
	// format to parse:
	// 127.0.0.1
	// 127.0.0.2
	lines, err := h.getIPSetEntries(name)
	if err != nil {
		return nil, err
	}
	result := make([]net.IP, 0, len(lines))
	for _, line := range lines {
		ip := net.ParseIP(strings.TrimSpace(line))
		if ip != nil {
			result = append(result, ip)
		}
	}
	return result, nil
}

func (h *execIPSetHelper) GetNetPorts(name string) ([]NetPort, error) {
	// format to parse:
	// 10.0.0.0/8,tcp:80
	lines, err := h.getIPSetEntries(name)
	if err != nil {
		return nil, err
	}
	result := make([]NetPort, 0, len(lines))
	for _, line := range lines {
		entry := strings.Split(strings.TrimSpace(line), ",")
		_, ipnet, err := net.ParseCIDR(entry[0])
		if err != nil {
			return nil, err
		}

		protoPort := strings.Split(entry[1], ":")
		proto, err := ParseProtocol(protoPort[0])
		if err != nil {
			return nil, err
		}

		pt, err := strconv.ParseUint(protoPort[1], 10, 16)
		if err != nil {
			return nil, err
		}
		port := uint16(pt)

		result = append(result, NetPort{
			Net:      *ipnet,
			Protocol: proto,
			Port:     port,
		})
	}
	return result, nil
}

func (h *execIPSetHelper) ensureSetHasOnlyGeneric(setName, typeName string, required []interface{},
	getter func(setName string) ([]interface{}, error),
	comparer func(e1, e2 interface{}) bool,
	adder func(setName string, obj interface{}) error,
	remover func(setName string, obj interface{}) error) error {

	current, err := getter(setName)
	if err != nil {
		return err
	}
	// find the diff
	toAdd, toRemove := collections.GetSlicesDifferences(required, current, comparer)

	for _, el := range toAdd {
		log.Debugf("Adding %s %v to ipset %s", typeName, el, setName)
		if err := adder(setName, el); err != nil {
			log.Errorf("Error adding entry %v to ipset %s", el, setName)
			return err
		}
	}
	for _, el := range toRemove {
		log.Debugf("Removing %s %v from ipset %s", typeName, el, setName)
		if err := remover(setName, el); err != nil {
			log.Debugf("Error removing entry %v from ipset %s", el, setName)
			return err
		}
	}

	return nil
}

func (h *execIPSetHelper) getIPSetEntries(name string) ([]string, error) {
	// # ipset list myset | awk '$0 ~ "^Members:$" {found=1; ln=NR}; NR>ln && found == 1 {print $1}'
	cmd := fmt.Sprintf(IPSetListWithAwk, name)
	res := h.exec.RunCommand("sh", "-c", cmd)
	if res.Err != nil || res.ExitCode != 0 {
		log.Debugf("Problem listing ipset %s - probably it's OK and it just doesn't exist: "+
			"%v, stdErr: %s", name, res.Err, res.StdErr)
		return nil, res.Err
	}
	lines := strings.Split(res.StdOut, "\n")
	return lines, nil
}

func (h *execIPSetHelper) addElementToSet(setName, elementTypeName string, element fmt.Stringer) error {
	res := h.exec.RunCommand("ipset", "add", setName, element.String())
	if res.Err != nil || res.ExitCode != 0 {
		log.Errorf("Error adding %s %s to ipset %s: %v, stdErr: %s",
			elementTypeName, element.String(), setName, res.Err, res.StdErr)
		return res.Err
	}
	return nil
}

func (h *execIPSetHelper) removeElementFromSet(setName, elementTypeName string, element fmt.Stringer) error {
	res := h.exec.RunCommand("ipset", "del", setName, element.String())
	if res.Err != nil || res.ExitCode != 0 {
		log.Debugf("Error removing %s %s from ipset %s: %v, stdErr: %s",
			elementTypeName, element.String(), setName, res.Err, res.StdErr)
		return res.Err
	}
	return nil
}

func ipSliceToInterface(ips []net.IP) []interface{} {
	res := make([]interface{}, len(ips))
	for i, ip := range ips {
		res[i] = ip
	}
	return res
}

func netPortSliceToInterface(nps []NetPort) []interface{} {
	res := make([]interface{}, len(nps))
	for i, np := range nps {
		res[i] = np
	}
	return res
}
