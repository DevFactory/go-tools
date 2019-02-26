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

package nettools_test

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	nt "github.com/DevFactory/go-tools/pkg/nettools"
	"github.com/DevFactory/go-tools/pkg/nettools/mocks"
	"github.com/DevFactory/go-tools/pkg/nettools/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestFilteredInterfaces(t *testing.T) {
	ethRegExp, _ := regexp.Compile("^eth[0-9]+")
	noneRegExp, _ := regexp.Compile("^NONE")
	allRegExp, _ := regexp.Compile(".*")
	type args struct {
		nameRegExp *regexp.Regexp
		lister     nt.InterfaceLister
	}
	tests := []struct {
		name    string
		args    args
		want    []nt.Interface
		wantErr bool
	}{
		{
			name:    "Check if eth0 and eth2 match" + ethRegExp.String(),
			args:    args{ethRegExp, &mocks.MockInterfaceLister{}},
			want:    mocks.MockInterfacesEth,
			wantErr: false,
		},
		{
			name:    "Check if nothing matches " + noneRegExp.String(),
			args:    args{noneRegExp, &mocks.MockInterfaceLister{}},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Check if everything matches " + allRegExp.String(),
			args:    args{allRegExp, &mocks.MockInterfaceLister{}},
			want:    mocks.MockInterfaces,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nt.FilteredInterfaces(tt.args.nameRegExp, tt.args.lister)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilteredInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilteredInterfaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoesntCreateIfNamedRoutingTableAlreadyExists(t *testing.T) {
	tableName := "eth1"
	testMustExistNamedRoutingTable(t, tableName, []byte("1687754607 eth1\n"), false)
}

func TestMustExistNamedRoutingTables(t *testing.T) {
	testCount := 100
	for i := 0; i < testCount; i++ {
		t.Run(fmt.Sprintf("random name test %d", i), func(t *testing.T) {
			tableName := testhelpers.RandStringBytes(10)
			testMustExistNamedRoutingTable(t, tableName, []byte{}, true)
		})
	}
}

func testMustExistNamedRoutingTable(t *testing.T, tableName string, rtTablesReadResult []byte, expectCreated bool) {
	ioOp := testhelpers.GetMockIOOpProviderWithEmptyRTTablesFile(rtTablesReadResult, 1, expectCreated)
	var created bool
	created, err := nt.MustExistNamedRoutingTable(tableName, ioOp)
	if err != nil {
		t.Errorf("MustExistNamedRoutingTable() error = %v", err)
		return
	}
	assert.True(t, created == expectCreated)
	ioOp.AssertExpectations(t)
	if created == false {
		return
	}
	args := ioOp.Calls[1].Arguments
	entry := args.String(1)
	last_char := string(entry[len(entry)-1])
	assert.Equal(t, "\n", last_char)
	routeTableEntry := strings.Fields(entry)
	assert.Len(t, routeTableEntry, 2)
	assert.Equal(t, tableName, routeTableEntry[1])
	id, err := strconv.Atoi(routeTableEntry[0])
	tableID := uint32(id)
	assert.Nil(t, err)
	assert.True(t, tableID > 255)
	assert.True(t, tableID>>31 == 0)
}

func TestOffsetAddr(t *testing.T) {
	tests := []struct {
		name    string
		n       string
		offset  int32
		want    net.IP
		wantErr bool
	}{
		{
			name:    "zero offset",
			n:       "10.10.10.10/24",
			offset:  0,
			want:    net.ParseIP("10.10.10.0"),
			wantErr: false,
		},
		{
			name:    "positive offset",
			n:       "10.10.10.10/24",
			offset:  1,
			want:    net.ParseIP("10.10.10.1"),
			wantErr: false,
		},
		{
			name:    "positive offset with subnet address",
			n:       "10.10.10.0/24",
			offset:  1,
			want:    net.ParseIP("10.10.10.1"),
			wantErr: false,
		},
		{
			name:    "positive offset 2",
			n:       "10.237.94.16/24",
			offset:  1,
			want:    net.ParseIP("10.237.94.1"),
			wantErr: false,
		},
		{
			name:    "negative offset",
			n:       "10.10.10.10/24",
			offset:  -1,
			want:    net.ParseIP("10.10.10.254"),
			wantErr: false,
		},
		// {
		// 	name:    "error on IPv6",
		// 	n:       "::1",
		// 	offset:  -1,
		// 	want:    net.ParseIP("10.10.10.254"),
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, n, _ := net.ParseCIDR(tt.n)
			got, err := nt.OffsetAddr(*n, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("OffsetAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("OffsetAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}
