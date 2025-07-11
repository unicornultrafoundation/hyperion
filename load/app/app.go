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

package app

import (
	"github.com/0xsoniclabs/hyperion/driver/rpc"
	"github.com/ethereum/go-ethereum/core/types"
)

//go:generate mockgen -source app.go -destination app_mock.go -package app

type Application interface {
	// CreateUsers creates a number of users for this application that
	// can generate transactions for this application.
	CreateUsers(context AppContext, numUsers int) ([]User, error)

	// GetReceivedTransactions returns the total number of transactions
	// received by this application up to the current point in time.
	GetReceivedTransactions(rpcClient rpc.Client) (uint64, error)
}

// User produces a stream of transactions to Generate traffic on the chain.
// Implementations are not required to be thread-safe.
type User interface {
	GenerateTx() (*types.Transaction, error)
	GetSentTransactions() uint64
}
