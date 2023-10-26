// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhealth

import (
	"net/http"

	health "github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
)

type Handler http.Handler

func NewHandler(h health.IHealth, custom map[string]interface{}) Handler {
	return handlers.NewJSONHandlerFunc(h, custom)
}
