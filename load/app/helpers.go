// Copyright 2024 Fantom Foundation
// This file is part of Norma System Testing Infrastructure for Sonic.
//
// Norma is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Norma is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Norma. If not, see <http://www.gnu.org/licenses/>.

package app

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xsoniclabs/norma/driver/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func createTx(from *Account, toAddress common.Address, value *big.Int, data []byte, gasLimit uint64) (*types.Transaction, error) {
	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     from.getNextNonce(),
		GasFeeCap: new(big.Int).Mul(big.NewInt(10_000), big.NewInt(1e9)),
		GasTipCap: big.NewInt(0),
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     value,
		Data:      data,
	})
	return types.SignTx(tx, types.NewLondonSigner(from.chainID), from.privateKey)
}

// GetGasPrice obtains optimal gasPrice for regular transactions
func GetGasPrice(rpcClient rpc.Client) (*big.Int, error) {
	gasPrice, err := rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price; %v", err)
	}
	var regularPrice big.Int
	regularPrice.Mul(gasPrice, big.NewInt(2)) // lower gas price for regular txs (but more than suggested by Opera)
	return &regularPrice, nil
}

func getPriorityGasPrice(regularGasPrice *big.Int) *big.Int {
	var priorityPrice big.Int
	priorityPrice.Mul(regularGasPrice, big.NewInt(2)) // greater gas price for init
	return &priorityPrice
}

func reverseAddresses(in []common.Address) []common.Address {
	out := make([]common.Address, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = in[len(in)-1-i]
	}
	return out
}
