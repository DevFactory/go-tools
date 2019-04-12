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
	"strings"
)

// Protocol is the type to provide different protocol names constants.
type Protocol string

const (
	// Unknown is the Protocol returned for unknown protocols
	Unknown Protocol = "unknown"
	// TCP is the name of the tcp protocol, as used in go's net library
	TCP Protocol = "tcp"
	// UDP is the name of the udp protocol, as used in go's net library
	UDP Protocol = "udp"
)

// ParseProtocol parses string and returns no error and a known Protocol.
// If the protocol name is unknown, returns Unknown and error.
func ParseProtocol(proto string) (Protocol, error) {
	p := strings.ToLower(proto)
	switch p {
	case "tcp":
		return TCP, nil
	case "udp":
		return UDP, nil
	default:
		return Unknown, fmt.Errorf("Unknown protocol %s passed for parsing", proto)
	}
}
