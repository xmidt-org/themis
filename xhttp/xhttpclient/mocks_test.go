// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpclient

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type mockRoundTripper struct {
	mock.Mock
}

func (mrt *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	arguments := mrt.Called(r)
	first, _ := arguments.Get(0).(*http.Response)
	return first, arguments.Error(1)
}

func (mrt *mockRoundTripper) ExpectRoundTrip(r *http.Request) *mock.Call {
	return mrt.On("RoundTrip", r)
}
