package token

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequestClaimBuilder(t *testing.T) {
	testData := []struct {
		request  *Request
		expected map[string]interface{}
	}{
		{
			request:  new(Request),
			expected: map[string]interface{}{},
		},
		{
			request: &Request{
				Claims: map[string]interface{}{"foo": 1, "bar": "value"},
			},
			expected: map[string]interface{}{"foo": 1, "bar": "value"},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				actual = make(map[string]interface{})
			)

			assert.NoError(
				requestClaimBuilder{}.AddClaims(context.Background(), record.request, actual),
			)

			assert.Equal(record.expected, actual)
		})
	}
}

func TestStaticClaimBuilder(t *testing.T) {
	testData := []struct {
		builder  staticClaimBuilder
		expected map[string]interface{}
	}{
		{
			expected: map[string]interface{}{},
		},
		{
			builder:  staticClaimBuilder{},
			expected: map[string]interface{}{},
		},
		{
			builder:  staticClaimBuilder{"foo": 1, "bar": "value"},
			expected: map[string]interface{}{"foo": 1, "bar": "value"},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				actual = make(map[string]interface{})
			)

			assert.NoError(
				record.builder.AddClaims(context.Background(), new(Request), actual),
			)

			assert.Equal(record.expected, actual)
		})
	}
}

func TestTimeClaimBuilder(t *testing.T) {
	var (
		expectedNow = time.Now()
		now         = func() time.Time { return expectedNow }
		testData    = []struct {
			builder  timeClaimBuilder
			expected map[string]interface{}
		}{
			{
				builder: timeClaimBuilder{
					now:              now,
					disableNotBefore: true,
				},
				expected: map[string]interface{}{
					"iat": expectedNow.UTC().Unix(),
				},
			},
			{
				builder: timeClaimBuilder{
					now: now,
				},
				expected: map[string]interface{}{
					"iat": expectedNow.UTC().Unix(),
					"nbf": expectedNow.UTC().Unix(),
				},
			},
			{
				builder: timeClaimBuilder{
					now:            now,
					duration:       24 * time.Hour,
					notBeforeDelta: 5 * time.Minute,
				},
				expected: map[string]interface{}{
					"iat": expectedNow.UTC().Unix(),
					"nbf": expectedNow.UTC().Add(5 * time.Minute).Unix(),
					"exp": expectedNow.UTC().Add(24 * time.Hour).Unix(),
				},
			},
			{
				builder: timeClaimBuilder{
					now:              now,
					duration:         30 * time.Minute,
					disableNotBefore: true,
				},
				expected: map[string]interface{}{
					"iat": expectedNow.UTC().Unix(),
					"exp": expectedNow.UTC().Add(30 * time.Minute).Unix(),
				},
			},
		}
	)

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				actual = make(map[string]interface{})
			)

			assert.NoError(
				record.builder.AddClaims(context.Background(), new(Request), actual),
			)

			assert.Equal(record.expected, actual)
		})
	}
}
