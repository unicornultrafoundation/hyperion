// Copyright 2024 Fantom Foundation
// This file is part of Hyperion System Testing Infrastructure for Sonic.
//
// Hyperion is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Hyperion is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Hyperion. If not, see <http://www.gnu.org/licenses/>.

package nodemon

import (
	"testing"
	"time"

	"github.com/0xsoniclabs/hyperion/driver"
	"github.com/0xsoniclabs/hyperion/driver/docker"
	opera "github.com/0xsoniclabs/hyperion/driver/node"
)

func TestCanCollectCpuProfileDateFromOperaNode(t *testing.T) {
	docker, err := docker.NewClient()
	if err != nil {
		t.Fatalf("failed to create a docker client: %v", err)
	}
	t.Cleanup(func() {
		_ = docker.Close()
	})
	node, err := opera.StartOperaDockerNode(docker, nil, &opera.OperaNodeConfig{
		Label:         "test",
		Image:         driver.DefaultClientDockerImageName,
		NetworkConfig: &driver.NetworkConfig{Validators: driver.DefaultValidators},
	})
	if err != nil {
		t.Fatalf("failed to create an Opera node on Docker: %v", err)
	}
	t.Cleanup(func() {
		_ = node.Cleanup()
	})
	data, err := GetPprofData(node, time.Second)
	if err != nil {
		t.Errorf("failed to collect pprof data from node: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("fetched empty CPU profile")
	}
}
