// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhealth

import (
	health "github.com/InVisionApp/go-health"
	"go.uber.org/fx"
)

// NopCheckable is an ICheckable that always returns success.  This type is useful when an app only
// wants a health endpoint to indicate its own liveness, rather than checking any external dependencies.
type NopCheckable struct {
	Details interface{}
}

func (nc NopCheckable) Status() (interface{}, error) {
	return nc.Details, nil
}

type ApplyChecksIn struct {
	fx.In

	Health health.IHealth
}

// ApplyChecks is an uber/fx Invoke function that allows checks created outside the uber/fx App
// to be registered with a Health component.
func ApplyChecks(first *health.Config, rest ...*health.Config) func(ApplyChecksIn) error {
	return func(in ApplyChecksIn) error {
		if err := in.Health.AddCheck(first); err != nil {
			return err
		}

		if len(rest) > 0 {
			if err := in.Health.AddChecks(rest); err != nil {
				return err
			}
		}

		return nil
	}
}
