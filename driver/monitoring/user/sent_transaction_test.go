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

package user

import (
	"testing"

	"github.com/0xsoniclabs/hyperion/driver"
	"go.uber.org/mock/gomock"
)

func TestSentTransactionSensorReportsProperValue(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		account int
		count   uint64
	}{
		{0, 4},
		{1, 5},
		{2, 3},
		{4, 0},
	}

	factory := &sentTransactionsSensorFactory{}

	for _, test := range tests {
		application := driver.NewMockApplication(ctrl)
		application.EXPECT().GetSentTransactions(test.account).Return(test.count, nil)

		sensor, err := factory.CreateSensor(application, test.account)
		if err != nil {
			t.Fatalf("creation of sensor failed: %v", err)
		}
		if res, err := sensor.ReadValue(); err != nil || res != int(test.count) {
			t.Errorf("sensor fetched wrong value, wanted %d, got %d, err %v", test.count, res, err)
		}
	}

}
