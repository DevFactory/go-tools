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
	"net"
	"reflect"
	"testing"
)

func TestTryRemove(t *testing.T) {
	tests := []struct {
		name string
		set  []net.IP
		ip   net.IP
		want []net.IP
	}{
		{
			name: "remove from empty",
			set:  []net.IP{},
			ip:   net.ParseIP("127.0.0.1"),
			want: []net.IP{},
		},
		{
			name: "remove existing from 1 element set",
			set:  []net.IP{net.ParseIP("127.0.0.1")},
			ip:   net.ParseIP("127.0.0.1"),
			want: []net.IP{},
		},
		{
			name: "remove non existing from 1 element set",
			set:  []net.IP{net.ParseIP("127.0.0.1")},
			ip:   net.ParseIP("127.0.0.10"),
			want: []net.IP{net.ParseIP("127.0.0.1")},
		},
		{
			name: "remove non-existing from multi element set",
			set:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2"), net.ParseIP("127.0.0.3")},
			ip:   net.ParseIP("127.0.0.10"),
			want: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2"), net.ParseIP("127.0.0.3")},
		},
		{
			name: "remove 1st from multi element set",
			set:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2"), net.ParseIP("127.0.0.3")},
			ip:   net.ParseIP("127.0.0.1"),
			want: []net.IP{net.ParseIP("127.0.0.2"), net.ParseIP("127.0.0.3")},
		},
		{
			name: "remove last from multi element set",
			set:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2"), net.ParseIP("127.0.0.3")},
			ip:   net.ParseIP("127.0.0.3"),
			want: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2")},
		},
		{
			name: "remove inside multi element set",
			set:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2"), net.ParseIP("127.0.0.3")},
			ip:   net.ParseIP("127.0.0.2"),
			want: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.3")},
		},
		{
			name: "remove from set with copies",
			set:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.3")},
			ip:   net.ParseIP("127.0.0.1"),
			want: []net.IP{net.ParseIP("127.0.0.3")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TryRemove(tt.set, tt.ip); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TryRemove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name   string
		first  []net.IP
		second []net.IP
		want   []net.IP
	}{
		{
			name:   "merge two empty",
			first:  []net.IP{},
			second: []net.IP{},
			want:   []net.IP{},
		},
		{
			name:   "merge first empty",
			first:  []net.IP{},
			second: []net.IP{net.ParseIP("127.0.0.1")},
			want:   []net.IP{net.ParseIP("127.0.0.1")},
		},
		{
			name:   "merge second empty",
			first:  []net.IP{net.ParseIP("127.0.0.1")},
			second: []net.IP{},
			want:   []net.IP{net.ParseIP("127.0.0.1")},
		},
		{
			name:   "merge unique",
			first:  []net.IP{net.ParseIP("127.0.0.1")},
			second: []net.IP{net.ParseIP("127.0.0.2")},
			want:   []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2")},
		},
		{
			name:   "merge non unique",
			first:  []net.IP{net.ParseIP("127.0.0.1")},
			second: []net.IP{net.ParseIP("127.0.0.2"), net.ParseIP("127.0.0.2")},
			want:   []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("127.0.0.2")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Merge(tt.first, tt.second); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPSliceFromStrings(t *testing.T) {
	tests := []struct {
		name      string
		addresses []string
		want      []net.IP
	}{
		{
			name:      "parse correct",
			addresses: []string{"127.0.0.1", "10.0.0.0"},
			want:      []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("10.0.0.0")},
		},
		{
			name:      "parse empty",
			addresses: []string{},
			want:      []net.IP{},
		},
		{
			name:      "parse incorrect",
			addresses: []string{"1271.0.0.1", "256.0.0.0"},
			want:      []net.IP{nil, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPSliceFromStrings(tt.addresses); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IPSliceFromStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSliceFromIPs(t *testing.T) {
	tests := []struct {
		name      string
		addresses []net.IP
		want      []string
	}{
		{
			name:      "convert empty",
			addresses: []net.IP{},
			want:      []string{},
		},
		{
			name:      "convert correct",
			addresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("10.0.0.0")},
			want:      []string{"127.0.0.1", "10.0.0.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringSliceFromIPs(tt.addresses); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringSliceFromIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}
