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

import "net"

// TryRemove removes all occurrences of ip from set and returns
// the resulting slice
func TryRemove(set []net.IP, ip net.IP) []net.IP {
	result := set

	firstFoundIndex := -1
	for i, a := range result {
		if a.Equal(ip) {
			firstFoundIndex = i
			break
		}
	}
	if firstFoundIndex > -1 {
		oldResult := result
		result = make([]net.IP, 0, firstFoundIndex)
		if firstFoundIndex > 0 {
			result = append(result, oldResult[0:firstFoundIndex]...)
		}
		for i := firstFoundIndex + 1; i < len(oldResult); i++ {
			if !oldResult[i].Equal(ip) {
				result = append(result, oldResult[i])
			}
		}
	}
	return result
}

// ContainsIP checks if a IP slice contains the IP
func ContainsIP(slice []net.IP, s net.IP) bool {
	for _, item := range slice {
		if item.Equal(s) {
			return true
		}
	}
	return false
}

// Merge adds all elements from second not existing in first
// to first and returns the resulting slice
func Merge(first, second []net.IP) []net.IP {
	result := make([]net.IP, len(first))
	copy(result, first)
	for _, fromSecond := range second {
		if !ContainsIP(result, fromSecond) {
			result = append(result, fromSecond)
		}
	}
	return result
}

// IPSliceFromStrings takes a slice of strings representing IP addresses
// and converts them into a slice of net.IP
func IPSliceFromStrings(addresses []string) []net.IP {
	result := []net.IP{}
	for _, addr := range addresses {
		result = append(result, net.ParseIP(addr))
	}
	return result
}

// StringSliceFromIPs takes a slice of net.IP representing IP addresses
// and converts them into a slice of strings
func StringSliceFromIPs(addresses []net.IP) []string {
	result := []string{}
	for _, addr := range addresses {
		result = append(result, addr.String())
	}
	return result
}
