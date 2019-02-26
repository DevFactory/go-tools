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
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// FilteredInterfaces returns interfaces like net.Interfaces(), but only those
// whose names match the given nameRegExp
func FilteredInterfaces(nameRegExp *regexp.Regexp, lister InterfaceLister) ([]Interface, error) {
	allInterfaces, err := lister.GetInterfaces()
	if err != nil {
		return nil, err
	}
	var res []Interface
	for _, iface := range allInterfaces {
		log.Debugf("Checking if interface %s matches filter...", iface.GetName())
		if nameRegExp.MatchString(iface.GetName()) {
			log.Debugf("%s matches config filter, adding to serviced interfaces.", iface.GetName())
			res = append(res, iface)
		}
	}
	return res, nil
}

// SimpleFileOperator provides simple subset of typical text file
// IO operations
//
// ReadFile reads the contents of a whole file and returns it as []byte
//
// AppendToFile tries to appens string textToAppend at the end of file named
// fileName
type SimpleFileOperator interface {
	ReadFile(fileName string) ([]byte, error)
	AppendToFile(fileName, textToAppend string) error
}

// NewIOSimpleFileOperator returns new instance of SimpleFileOperator
// implemented with real IO operations from ioutil package
func NewIOSimpleFileOperator() SimpleFileOperator {
	return &ioSimpleFileOperator{}
}

type ioSimpleFileOperator struct{}

func (i *ioSimpleFileOperator) ReadFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fileName)
}

func (i *ioSimpleFileOperator) AppendToFile(fileName, textToAppend string) error {
	var f *os.File
	var err error
	if f, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		return err
	}
	if _, err = f.WriteString(textToAppend); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// RTTablesFilename contains path to the file with ID to routing tables IDs mapping in Linux
const RTTablesFilename = "/etc/iproute2/rt_tables"

// MustExistNamedRoutingTable gets a routing table name as string tableName
// and creates a 32 int ID for that routing table. The pair [ID, tableName]
// is then written to the /etc/iproute2/rt_tables file to make it possible
// to reference by name the routing table tableName in "ip rule" command.
// It returns true if an entry was added to the file, false otherwise or when error.
func MustExistNamedRoutingTable(tableName string, ioOp SimpleFileOperator) (bool, error) {
	hasher := md5.New()
	io.WriteString(hasher, tableName)
	hash := hasher.Sum(nil)[0:4]
	// rt_tables supports 32b IDs, but only if there are positive signed int32
	// make sure the highest bit is == 0
	hash[3] >>= 1
	// IDs in the range 0 <= ID <= 255 are used by the operating system
	// make sure we won't use this range
	hash[0] |= 1
	tableID := uint32(uint32(hash[3])<<24 + uint32(hash[2])<<16 + uint32(hash[1])<<8 + uint32(hash[0]))
	tableEntry := fmt.Sprintf("%d %s", tableID, tableName)

	buf, err := ioOp.ReadFile(RTTablesFilename)
	if err != nil {
		return false, err
	}
	fileText := string(buf)
	for _, line := range strings.Split(fileText, "\n") {
		if line == tableEntry {
			return false, nil
		}
	}
	if err = ioOp.AppendToFile(RTTablesFilename, fmt.Sprintf("%s\n", tableEntry)); err != nil {
		return false, err
	}

	return true, nil
}

// OffsetAddr returns new net.IP address that is an offset of the subnet address of n.
// If the offset is positive, the result is the subnet address of IP in n plus the offset.
// If the offset is negative, the result is the broadcast address of IP in n minus the offset.
// Offset equal 0 returns the subnet address of IP in n.
func OffsetAddr(n net.IPNet, offset int32) (net.IP, error) {
	if offset == 0 {
		return n.IP, nil
	}
	if n.IP.To4() == nil {
		return net.IP{}, fmt.Errorf("does not support IPv6 addresses")
	}
	ip := make(net.IP, len(n.IP.To4()))
	if offset > 0 {
		binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(n.IP.To4())&binary.BigEndian.Uint32(net.IP(n.Mask).To4())+uint32(offset))
	}
	if offset < 0 {
		binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(n.IP.To4())|^binary.BigEndian.Uint32(net.IP(n.Mask).To4())-uint32(-offset))
	}
	return ip, nil
}
