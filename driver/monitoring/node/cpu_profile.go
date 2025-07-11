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
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/0xsoniclabs/hyperion/driver"
	"github.com/0xsoniclabs/hyperion/driver/monitoring"
	mon "github.com/0xsoniclabs/hyperion/driver/monitoring"
	"github.com/0xsoniclabs/hyperion/driver/monitoring/utils"
	opera "github.com/0xsoniclabs/hyperion/driver/node"
)

type PprofData []byte

func GetPprofData(node driver.Node, duration time.Duration) (PprofData, error) {
	url := node.GetServiceUrl(&opera.OperaDebugService)
	if url == nil {
		return nil, fmt.Errorf("node does not offer the pprof service")
	}
	resp, err := http.Get(fmt.Sprintf("%s/debug/pprof/profile?seconds=%d", *url, int(duration.Seconds())))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch result: %v", resp)
	}
	return io.ReadAll(resp.Body)
}

// NodeCpuProfile periodically collects CPU profiles from individual nodes.
var NodeCpuProfile = mon.Metric[mon.Node, mon.Series[mon.Time, string]]{
	Name:        "NodeCpuProfile",
	Description: "CpuProfile samples of a node at various times.",
}

func init() {
	if err := mon.RegisterSource(NodeCpuProfile, NewNodeCpuProfileSource); err != nil {
		panic(fmt.Sprintf("failed to register metric source: %v", err))
	}
}

// NewNodeCpuProfileSource creates a new data source periodically collecting
// CPU profiling data at configured sampling periods.
func NewNodeCpuProfileSource(monitor *monitoring.Monitor) mon.Source[mon.Node, mon.Series[mon.Time, string]] {
	return newPeriodicNodeDataSource[string](
		NodeCpuProfile,
		monitor,
		10*time.Second, // Sampling period; TODO: make customizable
		&cpuProfileSensorFactory{
			outputDir: monitor.Config().OutputDir,
		},
	)
}

type cpuProfileSensorFactory struct {
	outputDir string
}

func (f *cpuProfileSensorFactory) CreateSensor(node driver.Node) (utils.Sensor[string], error) {
	return &cpuProfileSensor{
		node:      node,
		duration:  5 * time.Second, // the duration of the CPU profile collection; TODO: make configurable
		outputDir: f.outputDir,
	}, nil
}

type cpuProfileSensor struct {
	node        driver.Node
	duration    time.Duration
	outputDir   string
	numProfiles int
}

func (s *cpuProfileSensor) ReadValue() (string, error) {
	data, err := GetPprofData(s.node, s.duration)
	if err != nil {
		return "", err
	}
	dir := fmt.Sprintf("%s/cpu_profiles/%s", s.outputDir, s.node.GetLabel())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%s/%06d.prof", dir, s.numProfiles)
	s.numProfiles++
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return "", err
	}
	return filename, nil
}
