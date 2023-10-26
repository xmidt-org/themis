// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

const (
	addressKey = "address"
	serverKey  = "server"
)

// AddressKey is the logging key for the server's bind address
func AddressKey() string {
	return addressKey
}

// ServerKey is the logging key for the server's name
func ServerKey() string {
	return serverKey
}
