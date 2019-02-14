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

package net

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// IPNetFromAddrWithMask converts from net.Addr with address in
// CIDR notation ("w.y.z.x/mm") into net.IPNet with IP and Mask
func IPNetFromAddrWithMask(addr net.Addr) (net.IPNet, error) {
	parsed := strings.Split(addr.String(), "/")
	if len(parsed) == 1 {
		addr := net.IPNet{
			IP:   net.ParseIP(parsed[0]),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		}
		return addr, nil
	}
	if len(parsed) == 2 {
		mask, err := strconv.Atoi(parsed[1])
		if err != nil {
			return net.IPNet{}, err
		}
		if mask < 0 || mask > 32 {
			return net.IPNet{}, fmt.Errorf("wrong IPv4 CIDR mask value: %v", mask)
		}
		binMask := net.CIDRMask(mask, 32)
		addr := net.IPNet{
			IP:   net.ParseIP(parsed[0]),
			Mask: binMask,
		}
		if addr.IP == nil {
			return net.IPNet{}, fmt.Errorf("wrong IPv4 address value: %v", parsed[0])
		}
		return addr, nil
	}

	return net.IPNet{}, fmt.Errorf("Can't parse %v into IP address and mask", addr)
}
