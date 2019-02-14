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

package strings

import (
	"reflect"
	"testing"
)

func TestCompareSlices(t *testing.T) {
	tests := []struct {
		name   string
		first  []string
		second []string
		want   bool
	}{
		{
			name:   "Empty Equal",
			first:  []string{},
			second: []string{},
			want:   true,
		},
		{
			name:   "Simple equal",
			first:  []string{"a", "b"},
			second: []string{"a", "b"},
			want:   true,
		},
		{
			name:   "Non equal",
			first:  []string{"a", "b"},
			second: []string{"c", "b"},
			want:   false,
		},
		{
			name:   "Non equal empty",
			first:  []string{"a", "b"},
			second: []string{},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareSlices(tt.first, tt.second); got != tt.want {
				t.Errorf("CompareSlices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveString(t *testing.T) {
	tests := []struct {
		name       string
		slice      []string
		s          string
		wantResult []string
	}{
		{
			name:       "remove the only one",
			slice:      []string{"the one"},
			s:          "the one",
			wantResult: nil,
		},
		{
			name:       "remove one",
			slice:      []string{"the one", "another"},
			s:          "the one",
			wantResult: []string{"another"},
		},
		{
			name:       "remove repeated",
			slice:      []string{"the one", "the one", "another"},
			s:          "the one",
			wantResult: []string{"another"},
		},
		{
			name:       "remove non existing",
			slice:      []string{"the one", "another"},
			s:          "the",
			wantResult: []string{"the one", "another"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := RemoveString(tt.slice, tt.s); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("RemoveString() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestGetCommon(t *testing.T) {
	tests := []struct {
		name   string
		first  []string
		second []string
		want   []string
	}{
		{
			name:   "common of empties",
			first:  nil,
			second: nil,
			want:   nil,
		},
		{
			name:   "common of empty and non empty",
			first:  []string{"test"},
			second: nil,
			want:   nil,
		},
		{
			name:   "common of distinct",
			first:  []string{"test1"},
			second: []string{"test2"},
			want:   nil,
		},
		{
			name:   "common when has common",
			first:  []string{"test1", "test"},
			second: []string{"test", "test2"},
			want:   []string{"test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCommon(tt.first, tt.second); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCommon() = %v, want %v", got, tt.want)
			}
		})
	}
}
