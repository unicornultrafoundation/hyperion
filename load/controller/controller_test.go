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

package controller_test

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0xsoniclabs/hyperion/driver"
	"github.com/0xsoniclabs/hyperion/driver/rpc"
	"github.com/0xsoniclabs/hyperion/load/app"
	"github.com/0xsoniclabs/hyperion/load/controller"
	"github.com/0xsoniclabs/hyperion/load/shaper"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/mock/gomock"
)

func TestLoadGeneration_CanRealizeConstantTrafficShape(t *testing.T) {

	rates := []int{
		10, 20, 50, 100, 200, 500, 1000, 2000, 5000,
	}

	for _, rate := range rates {
		t.Run(fmt.Sprintf("linear rate %v", rate), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			net := driver.NewMockNetwork(ctrl)
			rpcClient := rpc.NewMockClient(ctrl)
			application := app.NewMockApplication(ctrl)
			user := app.NewMockUser(ctrl)
			transaction := types.Transaction{}

			treasure, err := app.NewAccount(0, PrivateKey, nil, FakeNetworkID)
			if err != nil {
				t.Fatal(err)
			}

			check := NewRateCheck(float64(rate))
			var count atomic.Int32
			net.EXPECT().DialRandomRpc().AnyTimes().Return(rpcClient, nil)
			net.EXPECT().SendTransaction(gomock.Any()).AnyTimes().Do(func(any) {
				check.NewEvent()
				count.Add(1)
			})

			rpcClient.EXPECT().ChainID(gomock.Any()).Return(big.NewInt(0), nil).AnyTimes()
			rpcClient.EXPECT().NonceAt(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(0), nil)
			rpcClient.EXPECT().EstimateGas(gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(100), nil)
			rpcClient.EXPECT().SendTransaction(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			rpcClient.EXPECT().WaitTransactionReceipt(gomock.Any()).AnyTimes().Return(&types.Receipt{
				Status: types.ReceiptStatusSuccessful,
			}, nil)
			rpcClient.EXPECT().Close().AnyTimes().Return()

			users := make([]app.User, 100)
			for i := range users {
				users[i] = user
			}
			application.EXPECT().CreateUsers(gomock.Any(), 100).AnyTimes().Return(users, nil)

			rpcClient.EXPECT().SuggestGasPrice(gomock.Any()).AnyTimes().Return(big.NewInt(0), nil)
			user.EXPECT().GenerateTx().AnyTimes().Return(&transaction, nil)

			clientFactory := app.NewMockClientFactory(ctrl)
			clientFactory.EXPECT().DialRandomRpc().AnyTimes().Return(rpcClient, nil)

			shaper := shaper.NewConstantShaper(float64(rate))
			appContext, err := app.NewContext(clientFactory, treasure)
			if err != nil {
				t.Fatalf("failed to create app context: %v", err)
			}
			controller, err := controller.NewAppController(application, shaper, 100, appContext, net)
			if err != nil {
				t.Fatalf("failed to create app controller: %v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan bool)
			go func() {
				defer close(done)
				controller.Run(ctx)
			}()

			time.Sleep(time.Second)
			cancel()
			<-done

			// Check that the total number of processed messages is close to what is expected.
			got := float32(count.Load())
			want := float32(rate)
			if math.Abs(float64(got-want)) > math.Max(float64(want*0.02), 2.0) {
				t.Errorf("invalid number of produced messages, wanted ~%.0f, got %.0f", want, got)
			}

			// Check that during the execution the expected rate was within limits.
			if got, want := check.GetNumberOfUnderflows(), 0; got != want {
				t.Errorf("encountered %d times where messages have been produced too fast", got)
			}
			if got, want := check.GetNumberOfOverflows(), 0; got != want {
				t.Errorf("encountered %d times where messages have been produced too slow", got)
			}
		})
	}
}

type RateCheck struct {
	underflows atomic.Int32
	overflows  atomic.Int32
	mu         sync.Mutex
	level      float64
	last       time.Time
	rate       float64
	tolerance  float64
}

func NewRateCheck(rate float64) *RateCheck {
	return &RateCheck{
		rate:      rate,
		tolerance: math.Max(rate*0.1, 2.0),
	}
}

func (c *RateCheck) NewEvent() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	if c.last.IsZero() {
		c.last = now
		return
	}

	delta := now.Sub(c.last)

	c.level += delta.Seconds() * c.rate

	if c.level > c.tolerance {
		c.overflows.Add(1)
	}

	c.level -= 1
	if c.level < -c.tolerance {
		c.underflows.Add(1)
	}

	c.last = now
}

func (c *RateCheck) GetNumberOfUnderflows() int {
	return int(c.underflows.Load())
}

func (c *RateCheck) GetNumberOfOverflows() int {
	return int(c.overflows.Load())
}
