# Copyright 2025 SonicLabs
#
# This file is part of Hyperion System Testing Infrastructure for Sonic.
#
# Hyperion is free software: you can redistribute it and/or modify
# it under the terms of the GNU Lesser General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# Hyperion is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU lesser General Public License for more details.
#
# You should have received a copy of the GNU Lesser General Public License
# along with Hyperion. If not, see <http://www.gnu.org/licenses/>.

BUILD_DIR := $(CURDIR)/build

.PHONY: all test clean

all: genesistools

genesistools:
	go build -o $(BUILD_DIR)/genesistools ./app

test:
	go test ./... -v

clean:
	rm -rvf $(CURDIR)/build
