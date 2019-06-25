// Copyright 2019 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"config"
	"fmt"
	"os"

	"go.uber.org/fx"
)

const (
	applicationName    = "xmidt-issuer"
	applicationVersion = "0.0.0"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.FlagSet(config.FlagSetOptions{
				ApplicationName: applicationName,
				Arguments:       os.Args[1:],
			}),
			config.Viper(config.ViperOptions{
				ApplicationName: applicationName,
			}),
		),
	)

	if err := app.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start: %s", err)
		os.Exit(1)
	}

	app.Run()
}
