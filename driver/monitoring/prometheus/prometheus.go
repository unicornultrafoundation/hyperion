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

package prometheusmon

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0xsoniclabs/hyperion/driver"
	"github.com/0xsoniclabs/hyperion/driver/docker"
	"github.com/0xsoniclabs/hyperion/driver/network"
)

// PrometheusPort is the default port for the Prometheus service.
const PrometheusPort = 9090

// prometheusImage is the default Docker image for the Prometheus service.
const prometheusImage = "prom/prometheus:v2.44.0"

// Prometheus is a Prometheus instance running in a Docker container.
type Prometheus struct {
	container *docker.Container
	port      network.Port
	net       driver.Network
}

// Start starts a Prometheus instance in a Docker container.
func Start(net driver.Network, dn *docker.Network) (*Prometheus, error) {
	timeout := 1 * time.Second

	client, err := docker.NewClient()
	if err != nil {
		return nil, err
	}

	ports, err := network.GetFreePorts(1)
	if err != nil {
		return nil, err
	}

	// start the container
	container, err := client.Start(&docker.ContainerConfig{
		ImageName:       prometheusImage,
		ShutdownTimeout: &timeout,
		PortForwarding: map[network.Port]network.Port{
			PrometheusPort: ports[0],
		},
		Network: dn,
	})
	if err != nil {
		return nil, err
	}

	prometheus := &Prometheus{
		container: container,
		net:       net,
		port:      ports[0],
	}

	// initialize the config
	err = prometheus.initializeConfig()
	if err != nil {
		_ = container.Cleanup()
		return nil, err
	}

	// wait until the prometheus inside the Container is ready.
	// this is necessary for SIGHUP signal to be delivered correctly
	if err := network.Retry(network.DefaultRetryAttempts, 1*time.Second, func() error {
		resp, err := http.Get(prometheus.GetUrl() + "/-/ready")
		if err == nil && resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("not yet HTTP OK")
		}
		return err
	}); err == nil {
		log.Printf("started Prometheus on %s", prometheus.GetUrl())

		// listen for new Nodes
		net.RegisterListener(prometheus)

		// get nodes that have been started before this instance creation
		for _, node := range prometheus.net.GetActiveNodes() {
			prometheus.AfterNodeCreation(node)
		}

		return prometheus, nil
	}

	// if we reach this point, the prometheus instance is not ready
	_ = container.Cleanup()
	return nil, fmt.Errorf("prometheus instance is not ready")
}

// AddNode adds a new target to the Prometheus configuration to be observed.
func (p *Prometheus) AddNode(node driver.Node) error {
	cfg, err := renderConfigForNode(node)
	if err != nil {
		return err
	}
	_, err = p.container.Exec(
		[]string{"sh", "-c", fmt.Sprintf("echo '%s' > /etc/prometheus/opera-%s.json", cfg, node.Hostname())})
	if err != nil {
		return err
	}
	// we also need to reload the config
	return p.reloadConfig()
}

// Shutdown shuts down the Prometheus instance.
func (p *Prometheus) Shutdown() error {
	p.net.UnregisterListener(p)
	return p.container.Cleanup()
}

// GetUrl returns the URL of the Prometheus instance.
func (p *Prometheus) GetUrl() string {
	return fmt.Sprintf("http://localhost:%d", p.port)
}

func (p *Prometheus) AfterNodeCreation(node driver.Node) {
	if err := p.AddNode(node); err != nil {
		log.Printf("failed to add node %s to Prometheus: %s", node.Hostname(), err)
	}
}

func (p *Prometheus) AfterNodeRemoval(driver.Node) {
	// ignored
}

func (p *Prometheus) AfterApplicationCreation(driver.Application) {
	// ignored
}

// initializeConfig initializes the Prometheus configuration file by echoing config content
// into container's config location.
func (p *Prometheus) initializeConfig() error {
	_, err := p.container.Exec(
		[]string{"sh", "-c", fmt.Sprintf("echo '%s' > /etc/prometheus/prometheus.yml", promCfg)})
	return err
}

// reloadConfig reloads the Prometheus configuration by sending "SIGHUP" signal.
func (p *Prometheus) reloadConfig() error {
	return p.container.SendSignal(docker.SigHup)
}
