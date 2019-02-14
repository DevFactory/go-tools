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
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DevFactory/go-tools/pkg/linux/command"
	cmdmock "github.com/DevFactory/go-tools/pkg/linux/command/mock"
	"github.com/DevFactory/go-tools/pkg/nettools"
)

func TestConntrackHelperImpl_RemoveEntriesDNATingToIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      net.IP
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "check if conntrack command is passed OK",
			ip:      net.ParseIP("192.168.1.1"),
			want:    time.Second,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execRes := &cmdmock.ExecInfo{
				Expected: "conntrack -D -g -r 192.168.1.1",
				Returned: &command.ExecResult{
					Duration: time.Second,
					ExitCode: 0,
				},
			}
			executor := cmdmock.NewMockExecutorFromInfos(t, execRes)
			c := nettools.NewExecConntrackHelper(executor)
			got, err := c.RemoveEntriesDNATingToIP(tt.ip)

			assert.Equal(t, 1, executor.CallNum)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecConntrackHelper.RemoveEntriesDNATingToIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConntrackHelperImpl.RemoveEntriesDNATingToIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
