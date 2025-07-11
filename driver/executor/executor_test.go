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

package executor

import (
	"fmt"
	"github.com/0xsoniclabs/hyperion/driver/checking"
	"reflect"
	"syscall"
	"testing"

	"github.com/0xsoniclabs/hyperion/driver"
	"github.com/0xsoniclabs/hyperion/driver/parser"
	"go.uber.org/mock/gomock"
)

func TestExecutor_RunEmptyScenario(t *testing.T) {
	ctrl := gomock.NewController(t)
	clock := NewSimClock()
	net := driver.NewMockNetwork(ctrl)
	scenario := parser.Scenario{
		Name:       "Test",
		Duration:   10,
		Validators: []parser.Validator{{Name: "validator"}},
	}

	if err := Run(clock, net, &scenario, nil); err != nil {
		t.Errorf("failed to run empty scenario: %v", err)
	}
	want := Seconds(10)
	if got := clock.Now(); got < want {
		t.Errorf("scenario execution did not complete all steps, expected end time %v, got %v", want, got)
	}
}

func TestExecutor_RunSingleNodeScenario(t *testing.T) {

	clock := NewSimClock()
	scenario := parser.Scenario{
		Name:       "Test",
		Duration:   10,
		Validators: []parser.Validator{{Name: "validator"}},
		Nodes: []parser.Node{{
			Name:  "A",
			Start: New[float32](3),
			End:   New[float32](7),
		}},
	}

	ctrl := gomock.NewController(t)
	net := driver.NewMockNetwork(ctrl)
	node := driver.NewMockNode(ctrl)

	// In this scenario, a node is expected to be created and shut down.
	gomock.InOrder(
		net.EXPECT().CreateNode(gomock.Any()).Return(node, nil),
		net.EXPECT().RemoveNode(node),
		node.EXPECT().Stop(),
		node.EXPECT().Cleanup(),
	)

	if err := Run(clock, net, &scenario, nil); err != nil {
		t.Errorf("failed to run scenario: %v", err)
	}
	want := Seconds(10)
	if got := clock.Now(); got < want {
		t.Errorf("scenario execution did not complete all steps, expected end time %v, got %v", want, got)
	}
}

func TestExecutor_RunMultipleNodeScenario(t *testing.T) {

	clock := NewSimClock()
	scenario := parser.Scenario{
		Name:       "Test",
		Duration:   10,
		Validators: []parser.Validator{{Name: "validator"}},
		Nodes: []parser.Node{{
			Name:      "A",
			Instances: New(2),
			Start:     New[float32](3),
			End:       New[float32](7),
		}},
	}

	ctrl := gomock.NewController(t)
	net := driver.NewMockNetwork(ctrl)
	node1 := driver.NewMockNode(ctrl)
	node2 := driver.NewMockNode(ctrl)

	// In this scenario, two nodes are created and stopped.
	gomock.InOrder(
		net.EXPECT().CreateNode(gomock.Any()).Return(node1, nil),
		net.EXPECT().RemoveNode(newIs(node1)),
		node1.EXPECT().Stop(),
		node1.EXPECT().Cleanup(),
	)
	gomock.InOrder(
		net.EXPECT().CreateNode(gomock.Any()).Return(node2, nil),
		net.EXPECT().RemoveNode(newIs(node2)),
		node2.EXPECT().Stop(),
		node2.EXPECT().Cleanup(),
	)

	if err := Run(clock, net, &scenario, nil); err != nil {
		t.Errorf("failed to run scenario: %v", err)
	}
	want := Seconds(10)
	if got := clock.Now(); got < want {
		t.Errorf("scenario execution did not complete all steps, expected end time %v, got %v", want, got)
	}
}

func TestExecutor_RunSingleApplicationScenario(t *testing.T) {

	clock := NewSimClock()
	scenario := parser.Scenario{
		Name:       "Test",
		Duration:   10,
		Validators: []parser.Validator{{Name: "validator"}},
		Applications: []parser.Application{{
			Name:  "A",
			Type:  "counter",
			Start: New[float32](3),
			End:   New[float32](7),
			Rate:  parser.Rate{Constant: New[float32](10)},
		}},
	}

	ctrl := gomock.NewController(t)
	net := driver.NewMockNetwork(ctrl)
	app := driver.NewMockApplication(ctrl)

	// In this scenario, an application is expected to be created and shut down.
	net.EXPECT().CreateApplication(gomock.Any()).Return(app, nil)
	app.EXPECT().Start()
	app.EXPECT().Stop()

	if err := Run(clock, net, &scenario, nil); err != nil {
		t.Errorf("failed to run scenario: %v", err)
	}
	want := Seconds(10)
	if got := clock.Now(); got < want {
		t.Errorf("scenario execution did not complete all steps, expected end time %v, got %v", want, got)
	}
}

func TestExecutor_RunMultipleApplicationScenario(t *testing.T) {

	clock := NewSimClock()
	scenario := parser.Scenario{
		Name:       "Test",
		Duration:   10,
		Validators: []parser.Validator{{Name: "validator"}},
		Applications: []parser.Application{{
			Name:      "A",
			Type:      "counter",
			Instances: New(2),
			Start:     New[float32](3),
			End:       New[float32](7),
			Rate:      parser.Rate{Constant: New[float32](10)},
		}},
	}

	ctrl := gomock.NewController(t)
	net := driver.NewMockNetwork(ctrl)
	app1 := driver.NewMockApplication(ctrl)
	app2 := driver.NewMockApplication(ctrl)

	// In this scenario, an application is expected to be created and shut down.
	net.EXPECT().CreateApplication(gomock.Any()).Return(app1, nil)
	net.EXPECT().CreateApplication(gomock.Any()).Return(app2, nil)
	app1.EXPECT().Start()
	app1.EXPECT().Stop()
	app2.EXPECT().Start()
	app2.EXPECT().Stop()

	if err := Run(clock, net, &scenario, nil); err != nil {
		t.Errorf("failed to run scenario: %v", err)
	}
	want := Seconds(10)
	if got := clock.Now(); got < want {
		t.Errorf("scenario execution did not complete all steps, expected end time %v, got %v", want, got)
	}
}

func TestExecutor_TestUserAbort(t *testing.T) {

	clock := NewWallTimeClock()
	scenario := parser.Scenario{
		Name:       "Test",
		Duration:   5,
		Validators: []parser.Validator{{Name: "validator"}},
		Nodes: []parser.Node{{
			Name:  "A",
			Start: New[float32](1),
			End:   New[float32](3),
		}},
	}

	ctrl := gomock.NewController(t)
	net := driver.NewMockNetwork(ctrl)
	node := driver.NewMockNode(ctrl)

	// In this scenario, a node is created, after which a user interrupt is send.
	net.EXPECT().CreateNode(gomock.Any()).Do(func(_ any) {
		fmt.Printf("Sending interrupt signal to local process ..\n")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}).Return(node, nil)

	checks := checking.InitNetworkChecks(net, nil)
	if err := Run(clock, net, &scenario, checks); err == nil {
		t.Errorf("a user interrupt error should be reported")
	}
	want := Seconds(1)
	if got := clock.Now(); got < want || got > want+Seconds(1) {
		t.Errorf("scenario execution did not complete on user interrupt, expected end time %v, got %v", want, got)
	}
}

func TestExecutor_scheduleNetworkRulesEvents(t *testing.T) {
	clock := NewSimClock()
	scenario := parser.Scenario{
		Name:     "Test",
		Duration: 10,
		NetworkRules: parser.NetworkRules{
			Updates: []parser.NetworkRulesUpdate{
				{Time: 2, Rules: map[string]string{"MAX_BLOCK_GAS": "20500000000"}},
				{Time: 6, Rules: map[string]string{"MAX_EPOCH_GAS": "1500000000000"}},
			},
		},
	}

	ctrl := gomock.NewController(t)
	net := driver.NewMockNetwork(ctrl)
	gomock.InOrder(
		net.EXPECT().ApplyNetworkRules(map[string]string{"MAX_BLOCK_GAS": "20500000000"}),
		net.EXPECT().ApplyNetworkRules(map[string]string{"MAX_EPOCH_GAS": "1500000000000"}),
	)

	if err := Run(clock, net, &scenario, nil); err != nil {
		t.Errorf("failed to run scenario: %v", err)
	}
}

func New[T any](value T) *T {
	res := new(T)
	*res = value
	return res
}

type is[T any] struct {
	x T
}

func (e *is[T]) Matches(a any) bool {
	x, ok := a.(T)
	return ok && reflect.ValueOf(e.x) == reflect.ValueOf(x)
}

func (e *is[T]) String() string {
	return fmt.Sprintf("is %v", e.x)
}

func newIs[T any](node T) *is[T] {
	return &is[T]{node}
}
