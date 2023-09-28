// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package randomtest

import (
	"github.com/stretchr/testify/mock"
)

// Noncer is a stretchr mock for the random.Noncer interface.
type Noncer struct {
	mock.Mock
}

func (n *Noncer) Nonce() (string, error) {
	arguments := n.Called()
	return arguments.String(0), arguments.Error(1)
}

// Expect sets an expectation for a Nonce() call.  The returned Call object
// is returned so that it can be further customized.
func (n *Noncer) ExpectNonce() *mock.Call {
	return n.On("Nonce")
}
