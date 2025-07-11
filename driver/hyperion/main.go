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

package main

import (
	"fmt"
	"os"

	"github.com/0xsoniclabs/hyperion/driver/globalflags"
	"github.com/urfave/cli/v2"
)

// Run with `go run ./driver/hyperion`

func main() {
	app := &cli.App{
		Name:      "Hyperion Network Runner",
		HelpName:  "hyperion",
		Usage:     "A set of tools for running network scenarios",
		Copyright: "(c) 2023 Fantom Foundation",
		Flags:     globalflags.AllGlobalFlags,
		Commands: []*cli.Command{
			&checkCommand,
			&runCommand,
			&purgeCommand,
			&renderCommand,
			&diffCommand,
		},
		Before: globalflags.ProcessGlobalFlags,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
