// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpclient

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testChainThenNilRoundTripper(t *testing.T) {
	assert := assert.New(t)
	assert.IsType((*http.Transport)(nil), NewChain().Then(nil))
}

func testChainThenConstructors(t *testing.T, constructorCount int) {
	var (
		assert = assert.New(t)

		roundTripper     = new(mockRoundTripper)
		expectedRequest  = new(http.Request)
		expectedResponse = new(http.Response)
		expectedErr      = errors.New("expected")

		actualOrder  []int
		constructors []Constructor
	)

	for i := 0; i < constructorCount; i++ {
		constructors = append(constructors, func(i int) Constructor {
			return func(next http.RoundTripper) http.RoundTripper {
				return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
					assert.Equal(expectedRequest, r)
					actualOrder = append(actualOrder, i)
					return next.RoundTrip(r)
				})
			}
		}(i))
	}

	var expectedOrder []int
	for i := 0; i < constructorCount; i++ {
		expectedOrder = append(expectedOrder, i)
	}

	roundTripper.ExpectRoundTrip(expectedRequest).Once().Return(expectedResponse, expectedErr)
	actualResponse, actualErr := NewChain(constructors...).Then(roundTripper).RoundTrip(expectedRequest) //nolint: bodyclose
	assert.Equal(expectedResponse, actualResponse)
	assert.Equal(expectedErr, actualErr)
	assert.Equal(expectedOrder, actualOrder)
	roundTripper.AssertExpectations(t)
}

func testChainThen(t *testing.T) {
	t.Run("NilRoundTripper", testChainThenNilRoundTripper)

	t.Run("Constructors", func(t *testing.T) {
		for _, count := range []int{0, 1, 2, 3} {
			t.Run(fmt.Sprintf("count=%d", count), func(t *testing.T) {
				testChainThenConstructors(t, count)
			})
		}
	})
}

func testChainThenFuncNilRoundTripper(t *testing.T) {
	assert := assert.New(t)
	assert.IsType((*http.Transport)(nil), NewChain().ThenFunc(nil))
}

func testChainThenFuncConstructors(t *testing.T, constructorCount int) {
	var (
		assert = assert.New(t)

		roundTripper     = new(mockRoundTripper)
		expectedRequest  = new(http.Request)
		expectedResponse = new(http.Response)
		expectedErr      = errors.New("expected")

		actualOrder  []int
		constructors []Constructor
	)

	for i := 0; i < constructorCount; i++ {
		constructors = append(constructors, func(i int) Constructor {
			return func(next http.RoundTripper) http.RoundTripper {
				return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
					assert.Equal(expectedRequest, r)
					actualOrder = append(actualOrder, i)
					return next.RoundTrip(r)
				})
			}
		}(i))
	}

	var expectedOrder []int
	for i := 0; i < constructorCount; i++ {
		expectedOrder = append(expectedOrder, i)
	}

	roundTripper.ExpectRoundTrip(expectedRequest).Once().Return(expectedResponse, expectedErr)
	actualResponse, actualErr := NewChain(constructors...).ThenFunc(roundTripper.RoundTrip).RoundTrip(expectedRequest) //nolint: bodyclose
	assert.Equal(expectedResponse, actualResponse)
	assert.Equal(expectedErr, actualErr)
	assert.Equal(expectedOrder, actualOrder)
	roundTripper.AssertExpectations(t)
}

func testChainThenFunc(t *testing.T) {
	t.Run("NilRoundTripper", testChainThenFuncNilRoundTripper)

	t.Run("Constructors", func(t *testing.T) {
		for _, count := range []int{0, 1, 2, 3} {
			t.Run(fmt.Sprintf("count=%d", count), func(t *testing.T) {
				testChainThenFuncConstructors(t, count)
			})
		}
	})
}

func testChainAppend(t *testing.T) {
	for _, initialCount := range []int{0, 1, 2, 3} {
		t.Run(fmt.Sprintf("initialCount=%d", initialCount), func(t *testing.T) {
			for _, appendCount := range []int{0, 1, 2, 3} {
				t.Run(fmt.Sprintf("appendCount=%d", appendCount), func(t *testing.T) {
					var (
						assert = assert.New(t)

						roundTripper     = new(mockRoundTripper)
						expectedRequest  = new(http.Request)
						expectedResponse = new(http.Response)
						expectedErr      = errors.New("expected")

						actualOrder []int
						initial     []Constructor
						more        []Constructor
					)

					for i := 0; i < initialCount; i++ {
						initial = append(initial, func(i int) Constructor {
							return func(next http.RoundTripper) http.RoundTripper {
								return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
									assert.Equal(expectedRequest, r)
									actualOrder = append(actualOrder, i)
									return next.RoundTrip(r)
								})
							}
						}(i))
					}

					for i := 0; i < appendCount; i++ {
						more = append(more, func(i int) Constructor {
							return func(next http.RoundTripper) http.RoundTripper {
								return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
									assert.Equal(expectedRequest, r)
									actualOrder = append(actualOrder, i)
									return next.RoundTrip(r)
								})
							}
						}(i+initialCount))
					}

					var expectedOrder []int
					for i := 0; i < (initialCount + appendCount); i++ {
						expectedOrder = append(expectedOrder, i)
					}

					roundTripper.ExpectRoundTrip(expectedRequest).Once().Return(expectedResponse, expectedErr)
					actualResponse, actualErr := NewChain(initial...).Append(more...).Then(roundTripper).RoundTrip(expectedRequest) //nolint: bodyclose
					assert.Equal(expectedResponse, actualResponse)                                                                  //nolint: bodyclose
					assert.Equal(expectedErr, actualErr)
					assert.Equal(expectedOrder, actualOrder)
					roundTripper.AssertExpectations(t)
				})
			}
		})
	}
}

func testChainExtend(t *testing.T) {
	for _, initialCount := range []int{0, 1, 2, 3} {
		t.Run(fmt.Sprintf("initialCount=%d", initialCount), func(t *testing.T) {
			for _, appendCount := range []int{0, 1, 2, 3} {
				t.Run(fmt.Sprintf("appendCount=%d", appendCount), func(t *testing.T) {
					var (
						assert = assert.New(t)

						roundTripper     = new(mockRoundTripper)
						expectedRequest  = new(http.Request)
						expectedResponse = new(http.Response)
						expectedErr      = errors.New("expected")

						actualOrder []int
						initial     []Constructor
						more        []Constructor
					)

					for i := 0; i < initialCount; i++ {
						initial = append(initial, func(i int) Constructor {
							return func(next http.RoundTripper) http.RoundTripper {
								return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
									assert.Equal(expectedRequest, r)
									actualOrder = append(actualOrder, i)
									return next.RoundTrip(r)
								})
							}
						}(i))
					}

					for i := 0; i < appendCount; i++ {
						more = append(more, func(i int) Constructor {
							return func(next http.RoundTripper) http.RoundTripper {
								return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
									assert.Equal(expectedRequest, r)
									actualOrder = append(actualOrder, i)
									return next.RoundTrip(r)
								})
							}
						}(i+initialCount))
					}

					var expectedOrder []int
					for i := 0; i < (initialCount + appendCount); i++ {
						expectedOrder = append(expectedOrder, i)
					}

					roundTripper.ExpectRoundTrip(expectedRequest).Once().Return(expectedResponse, expectedErr)
					actualResponse, actualErr := NewChain(initial...).Extend(NewChain(more...)).Then(roundTripper).RoundTrip(expectedRequest) //nolint: bodyclose
					assert.Equal(expectedResponse, actualResponse)                                                                            //nolint: bodyclose
					assert.Equal(expectedErr, actualErr)
					assert.Equal(expectedOrder, actualOrder)
					roundTripper.AssertExpectations(t)
				})
			}
		})
	}
}

func TestChain(t *testing.T) {
	t.Run("Then", testChainThen)
	t.Run("ThenFunc", testChainThenFunc)
	t.Run("Append", testChainAppend)
	t.Run("Extend", testChainExtend)
}
