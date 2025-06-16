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

package external

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xsoniclabs/hyperion/driver"
	"github.com/0xsoniclabs/hyperion/driver/network"
	rpcdriver "github.com/0xsoniclabs/hyperion/driver/rpc"
	"github.com/0xsoniclabs/hyperion/load/app"
	"github.com/0xsoniclabs/hyperion/load/controller"
	"github.com/0xsoniclabs/hyperion/load/shaper"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// ExternalNetwork connects to existing external blockchain nodes
// instead of creating new Docker containers
type ExternalNetwork struct {
	config         driver.NetworkConfig
	primaryAccount *app.Account
	rpcEndpoints   []string // List of RPC endpoints to connect to

	// apps maintains a list of applications
	apps      []driver.Application
	appsMutex sync.Mutex
	nextAppId atomic.Uint32

	// listeners for network events
	listeners     map[driver.NetworkListener]bool
	listenerMutex sync.Mutex

	// app context for managing applications
	appContext app.AppContext
}

// ExternalNetworkConfig contains configuration for external network
type ExternalNetworkConfig struct {
	NetworkConfig driver.NetworkConfig
	RpcEndpoints  []string // HTTP RPC endpoints (e.g., ["http://localhost:18545"])
	ChainID       int64    // Chain ID of your network
}

// NewExternalNetwork creates a new network that connects to external nodes
func NewExternalNetwork(config *ExternalNetworkConfig) (*ExternalNetwork, error) {
	if len(config.RpcEndpoints) == 0 {
		return nil, fmt.Errorf("at least one RPC endpoint must be provided")
	}

	// Create primary account for the network operations
	// You may need to adjust this private key to one that has funds on your chain
	primaryAccount, err := app.NewAccount(0, "163f5f0f9a621d72fedd85ffca3d08d131ab4e812181e0d30ffd1c885d20aac7", nil, config.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary account: %w", err)
	}

	net := &ExternalNetwork{
		config:         config.NetworkConfig,
		primaryAccount: primaryAccount,
		rpcEndpoints:   config.RpcEndpoints,
		apps:           []driver.Application{},
		listeners:      map[driver.NetworkListener]bool{},
	}

	// Setup app context for managing applications
	appContext, err := app.NewContext(net, primaryAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create app context: %w", err)
	}
	net.appContext = appContext

	return net, nil
}

// CreateNode - Not supported for external networks
func (n *ExternalNetwork) CreateNode(config *driver.NodeConfig) (driver.Node, error) {
	return nil, fmt.Errorf("creating nodes is not supported for external networks")
}

// RemoveNode - Not supported for external networks
func (n *ExternalNetwork) RemoveNode(node driver.Node) error {
	return fmt.Errorf("removing nodes is not supported for external networks")
}

// GetActiveNodes returns empty list since we don't manage nodes
func (n *ExternalNetwork) GetActiveNodes() []driver.Node {
	return []driver.Node{}
}

// CreateApplication creates applications that will send transactions to external chain
func (n *ExternalNetwork) CreateApplication(config *driver.ApplicationConfig) (driver.Application, error) {
	appId := n.nextAppId.Add(1)
	application, err := app.NewApplication(config.Type, n.appContext, 0, appId)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application: %w", err)
	}

	sh, err := shaper.ParseRate(config.Rate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rate: %w", err)
	}

	appController, err := controller.NewAppController(application, sh, config.Users, n.appContext, n)
	if err != nil {
		return nil, err
	}

	app := &externalApplication{
		name:       config.Name,
		controller: appController,
		config:     config,
		done:       &sync.WaitGroup{},
	}

	n.appsMutex.Lock()
	n.apps = append(n.apps, app)
	n.appsMutex.Unlock()

	n.listenerMutex.Lock()
	for listener := range n.listeners {
		listener.AfterApplicationCreation(app)
	}
	n.listenerMutex.Unlock()

	return app, nil
}

// GetActiveApplications returns list of active applications
func (n *ExternalNetwork) GetActiveApplications() []driver.Application {
	n.appsMutex.Lock()
	defer n.appsMutex.Unlock()
	return n.apps
}

// RegisterListener registers a network event listener
func (n *ExternalNetwork) RegisterListener(listener driver.NetworkListener) {
	n.listenerMutex.Lock()
	n.listeners[listener] = true
	n.listenerMutex.Unlock()
}

// UnregisterListener removes a network event listener
func (n *ExternalNetwork) UnregisterListener(listener driver.NetworkListener) {
	n.listenerMutex.Lock()
	delete(n.listeners, listener)
	n.listenerMutex.Unlock()
}

// SendTransaction sends a transaction to one of the external RPC endpoints
func (n *ExternalNetwork) SendTransaction(tx *types.Transaction) {
	// Send to the first available endpoint
	// You could implement load balancing here
	client, err := n.DialRandomRpc()
	if err != nil {
		log.Printf("failed to dial RPC: %v", err)
		return
	}
	defer client.Close()

	if err := client.SendTransaction(context.Background(), tx); err != nil {
		log.Printf("failed to send transaction: %v", err)
	}
}

// DialRandomRpc connects to one of the configured RPC endpoints
func (n *ExternalNetwork) DialRandomRpc() (rpcdriver.Client, error) {
	if len(n.rpcEndpoints) == 0 {
		return nil, fmt.Errorf("no RPC endpoints configured")
	}

	// For simplicity, use the first endpoint
	// You could implement random selection or load balancing
	endpoint := n.rpcEndpoints[0]

	rpcClient, err := network.RetryReturn(network.DefaultRetryAttempts, 1*time.Second, func() (*rpc.Client, error) {
		return rpc.DialContext(context.Background(), endpoint)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC endpoint %s: %w", endpoint, err)
	}

	return rpcdriver.WrapRpcClient(rpcClient), nil
}

// ApplyNetworkRules - Not supported for external networks since we don't control them
func (n *ExternalNetwork) ApplyNetworkRules(rules driver.NetworkRules) error {
	log.Printf("Warning: ApplyNetworkRules not supported for external networks")
	return nil
}

// Shutdown stops all applications and cleans up resources
func (n *ExternalNetwork) Shutdown() error {
	var errs []error

	// Stop all applications
	for _, app := range n.apps {
		if err := app.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	n.apps = n.apps[:0]

	// Close app context
	if n.appContext != nil {
		n.appContext.Close()
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

// externalApplication implements the Application interface
type externalApplication struct {
	name       string
	controller *controller.AppController
	config     *driver.ApplicationConfig
	cancel     context.CancelFunc
	done       *sync.WaitGroup
}

func (a *externalApplication) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	a.done.Add(1)
	go func() {
		defer a.done.Done()
		err := a.controller.Run(ctx)
		if err != nil {
			log.Printf("Application %s failed: %v", a.name, err)
		}
	}()
	return nil
}

func (a *externalApplication) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	a.cancel = nil
	log.Printf("Stopping application: %s", a.name)
	a.done.Wait()
	log.Printf("Application stopped: %s", a.name)
	return nil
}

func (a *externalApplication) Config() *driver.ApplicationConfig {
	return a.config
}

func (a *externalApplication) GetNumberOfUsers() int {
	return a.controller.GetNumberOfUsers()
}

func (a *externalApplication) GetSentTransactions(user int) (uint64, error) {
	return a.controller.GetTransactionsSentBy(user)
}

func (a *externalApplication) GetReceivedTransactions() (uint64, error) {
	return a.controller.GetReceivedTransactions()
}
