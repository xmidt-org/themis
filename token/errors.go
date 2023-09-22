// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

type httpError struct {
	err  error
	code int
}

func (e httpError) Error() string {
	return e.err.Error()
}

func (e httpError) StatusCode() int {
	return e.code
}
