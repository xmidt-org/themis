// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"bufio"
	"context"
	"net"
	"net/http"

	"github.com/stretchr/testify/mock"
)

type mockResponseWriter struct {
	mock.Mock
}

func (m *mockResponseWriter) Header() http.Header {
	var (
		arguments = m.Called()
		first, _  = arguments.Get(0).(http.Header)
	)

	return first
}

func (m *mockResponseWriter) ExpectHeader() *mock.Call {
	return m.On("Header")
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	arguments := m.Called(b)
	return arguments.Int(0), arguments.Error(1)
}

func (m *mockResponseWriter) ExpectWrite(b []byte) *mock.Call {
	return m.On("Write", b)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

func (m *mockResponseWriter) ExpectWriteHeader(statusCode int) *mock.Call {
	return m.On("WriteHeader", statusCode)
}

type mockHijacker struct {
	mock.Mock
}

func (m *mockHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	var (
		arguments = m.Called()
		first, _  = arguments.Get(0).(net.Conn)
		second, _ = arguments.Get(1).(*bufio.ReadWriter)
	)

	return first, second, arguments.Error(2)
}

func (m *mockHijacker) ExpectHijack() *mock.Call {
	return m.On("Hijack")
}

type mockFlusher struct {
	mock.Mock
}

func (m *mockFlusher) Flush() {
	m.Called()
}

func (m *mockFlusher) ExpectFlush() *mock.Call {
	return m.On("Flush")
}

type mockPusher struct {
	mock.Mock
}

func (m *mockPusher) Push(target string, opts *http.PushOptions) error {
	return m.Called(target, opts).Error(0)
}

func (m *mockPusher) ExpectPush(target string, opts *http.PushOptions) *mock.Call {
	return m.On("Push", target, opts)
}

type hijackerWriter struct {
	http.ResponseWriter
	http.Hijacker
}

type pusherWriter struct {
	http.ResponseWriter
	http.Pusher
}

type flusherWriter struct {
	http.ResponseWriter
	http.Flusher
}

type mockServer struct {
	mock.Mock
}

func (m *mockServer) Serve(l net.Listener) error {
	return m.Called(l).Error(0)
}

func (m *mockServer) ExpectServe(p ...interface{}) *mock.Call {
	return m.On("Serve", p...)
}

func (m *mockServer) Shutdown(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockServer) ExpectShutdown(p ...interface{}) *mock.Call {
	return m.On("Shutdown", p...)
}
