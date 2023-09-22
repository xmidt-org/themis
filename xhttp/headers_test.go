// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttp

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalizeHeaders(t *testing.T) {
	testData := []struct {
		source   http.Header
		expected http.Header
	}{
		{
			expected: http.Header{},
		},
		{
			source:   http.Header{},
			expected: http.Header{},
		},
		{
			source:   http.Header{"Content-Type": []string{"text/plain"}, "x-test": []string{"value1", "value2"}},
			expected: http.Header{"Content-Type": []string{"text/plain"}, "X-Test": []string{"value1", "value2"}},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				actual = CanonicalizeHeaders(record.source)
			)

			assert.Equal(record.expected, actual)
		})
	}
}

func TestCanonicalizeHeaderMap(t *testing.T) {
	testData := []struct {
		source   map[string]string
		expected http.Header
	}{
		{
			expected: http.Header{},
		},
		{
			source:   map[string]string{},
			expected: http.Header{},
		},
		{
			source:   map[string]string{"Content-Type": "text/plain", "x-test": "value"},
			expected: http.Header{"Content-Type": []string{"text/plain"}, "X-Test": []string{"value"}},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				actual = CanonicalizeHeaderMap(record.source)
			)

			assert.Equal(record.expected, actual)
		})
	}
}

func TestAddHeaders(t *testing.T) {
	testData := []struct {
		source   http.Header
		target   http.Header
		expected http.Header
	}{
		{
			target:   http.Header{},
			expected: http.Header{},
		},
		{
			source:   http.Header{},
			target:   http.Header{},
			expected: http.Header{},
		},
		{
			source:   http.Header{"x-test-1": []string{"value"}, "X-eXIsting": []string{"new value"}},
			target:   http.Header{"X-eXIsting": []string{"existing value"}},
			expected: http.Header{"x-test-1": []string{"value"}, "X-eXIsting": []string{"existing value", "new value"}},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert := assert.New(t)
			AddHeaders(record.target, record.source)
			assert.Equal(record.expected, record.target)
		})
	}
}

func TestSetHeaders(t *testing.T) {
	testData := []struct {
		source   http.Header
		target   http.Header
		expected http.Header
	}{
		{
			target:   http.Header{},
			expected: http.Header{},
		},
		{
			source:   http.Header{},
			target:   http.Header{},
			expected: http.Header{},
		},
		{
			source:   http.Header{"x-test-1": []string{"value"}, "X-eXIsting": []string{"value1", "value2"}},
			target:   http.Header{"X-eXIsting": []string{"target value"}},
			expected: http.Header{"x-test-1": []string{"value"}, "X-eXIsting": []string{"value1", "value2"}},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert := assert.New(t)
			SetHeaders(record.target, record.source)
			assert.Equal(record.expected, record.target)
		})
	}
}
