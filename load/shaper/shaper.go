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

package shaper

import (
	"fmt"
	"time"

	"github.com/0xsoniclabs/hyperion/driver/parser"
)

//go:generate mockgen -source shaper.go -destination shaper_mock.go -package shaper

// Shaper defines the shape of traffic to be produced by an application.
type Shaper interface {
	// Start notifies the shaper that processing is started at the given time
	// and provides a source for fetching load information.
	Start(time.Time, LoadInfoSource)

	// GetNumMessagesInInterval provides the number of messages to be produced
	// in the given time interval. The result is expected to be >= 0.
	GetNumMessagesInInterval(start time.Time, duration time.Duration) float64
}

// LoadInfoSource defines an interface for load-sensitive traffic shapes to
// request load state information.
type LoadInfoSource interface {
	GetSentTransactions() (uint64, error)
	GetReceivedTransactions() (uint64, error)
}

// ParseRate parses rate from the parser.
func ParseRate(rate *parser.Rate) (Shaper, error) {
	// return default constant shaper if rate is not specified
	if rate == nil {
		return NewConstantShaper(0), nil
	}

	if rate.Constant != nil {
		return NewConstantShaper(float64(*rate.Constant)), nil
	}
	if rate.Slope != nil {
		return NewSlopeShaper(float64(rate.Slope.Start), float64(rate.Slope.Increment)), nil
	}
	if rate.Auto != nil {
		increase := 1.0
		if rate.Auto.Increase != nil {
			increase = float64(*rate.Auto.Increase)
		}
		decrease := 0.2
		if rate.Auto.Decrease != nil {
			decrease = float64(*rate.Auto.Decrease)
		}
		return NewAutoShaper(increase, decrease), nil
	}
	if rate.Wave != nil {
		min := float32(0)
		if rate.Wave.Min != nil {
			min = *rate.Wave.Min
		}
		return NewWaveShaper(min, rate.Wave.Max, rate.Wave.Period), nil
	}

	return nil, fmt.Errorf("unknown rate type")
}
