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

package collections

import (
	"reflect"
	"testing"
)

func TestGetSlicesDifferences(t *testing.T) {
	type args struct {
		first        []interface{}
		second       []interface{}
		equalityFunc func(first, second interface{}) bool
	}
	tests := []struct {
		name             string
		args             args
		wantInFirstOnly  []interface{}
		wantInSecondOnly []interface{}
	}{
		{
			name: "two empty sets",
			args: args{
				first:        []interface{}{},
				second:       []interface{}{},
				equalityFunc: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			wantInFirstOnly:  []interface{}{},
			wantInSecondOnly: []interface{}{},
		},
		{
			name: "two equal sets",
			args: args{
				first:        []interface{}{1, 2},
				second:       []interface{}{1, 2},
				equalityFunc: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			wantInFirstOnly:  []interface{}{},
			wantInSecondOnly: []interface{}{},
		},
		{
			name: "first empty set",
			args: args{
				first:        []interface{}{},
				second:       []interface{}{1, 2, 3},
				equalityFunc: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			wantInFirstOnly:  []interface{}{},
			wantInSecondOnly: []interface{}{1, 2, 3},
		},
		{
			name: "second empty set",
			args: args{
				first:        []interface{}{1, 2, 3},
				second:       []interface{}{},
				equalityFunc: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			wantInFirstOnly:  []interface{}{1, 2, 3},
			wantInSecondOnly: []interface{}{},
		},
		{
			name: "two partially overlapping sets",
			args: args{
				first:        []interface{}{1, 2, 3},
				second:       []interface{}{3, 4, 5},
				equalityFunc: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			wantInFirstOnly:  []interface{}{1, 2},
			wantInSecondOnly: []interface{}{4, 5},
		},
		{
			name: "two disjoined sets",
			args: args{
				first:        []interface{}{1, 2, 3},
				second:       []interface{}{4, 5},
				equalityFunc: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			wantInFirstOnly:  []interface{}{1, 2, 3},
			wantInSecondOnly: []interface{}{4, 5},
		},
		{
			name: "first included in second",
			args: args{
				first:        []interface{}{1, 2},
				second:       []interface{}{1, 2, 3},
				equalityFunc: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			wantInFirstOnly:  []interface{}{},
			wantInSecondOnly: []interface{}{3},
		},
		{
			name: "second included in first",
			args: args{
				first:        []interface{}{1, 2, 3},
				second:       []interface{}{1, 2},
				equalityFunc: func(a, b interface{}) bool { return a.(int) == b.(int) },
			},
			wantInFirstOnly:  []interface{}{3},
			wantInSecondOnly: []interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInFirstOnly, gotInSecondOnly := GetSlicesDifferences(tt.args.first, tt.args.second, tt.args.equalityFunc)
			if !reflect.DeepEqual(gotInFirstOnly, tt.wantInFirstOnly) {
				t.Errorf("GetSlicesDifferences() gotInFirstOnly = %v, want %v", gotInFirstOnly, tt.wantInFirstOnly)
			}
			if !reflect.DeepEqual(gotInSecondOnly, tt.wantInSecondOnly) {
				t.Errorf("GetSlicesDifferences() gotInSecondOnly = %v, want %v", gotInSecondOnly, tt.wantInSecondOnly)
			}
		})
	}
}
