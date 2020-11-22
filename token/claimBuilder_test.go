package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/xmidt-org/themis/random/randomtest"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"

	"github.com/stretchr/testify/suite"
)

type ClaimBuildersTestSuite struct {
	suite.Suite
	expectedCtx context.Context
	expectedErr error
}

var _ suite.SetupAllSuite = (*ClaimBuildersTestSuite)(nil)

func (suite *ClaimBuildersTestSuite) SetupSuite() {
	suite.expectedCtx = context.WithValue(context.Background(), "foo", "bar")
	suite.expectedErr = errors.New("expected AddClaims error")
}

func (suite *ClaimBuildersTestSuite) TestSuccess() {
	for _, count := range []int{0, 1, 2, 5} {
		suite.Run(fmt.Sprintf("count=%d", count), func() {
			var (
				builder         ClaimBuilders
				expectedRequest = new(Request)
				expected        = make(map[string]interface{})
				actual          = make(map[string]interface{})
			)

			for i := 0; i < count; i++ {
				i := i
				expected[strconv.Itoa(i)] = "true"
				builder = append(builder,
					ClaimBuilderFunc(func(actualCtx context.Context, actualRequest *Request, target map[string]interface{}) error {
						suite.Equal(suite.expectedCtx, actualCtx)
						suite.True(expectedRequest == actualRequest)
						target[strconv.Itoa(i)] = "true"
						return nil
					}),
				)
			}

			suite.Require().NoError(
				builder.AddClaims(suite.expectedCtx, expectedRequest, actual),
			)

			suite.Equal(expected, actual)
		})
	}
}

func (suite *ClaimBuildersTestSuite) TestError() {
	var (
		expectedRequest = new(Request)
		expected        = map[string]interface{}{
			"first": "true",
		}

		actual = make(map[string]interface{})

		builder = ClaimBuilders{
			ClaimBuilderFunc(func(actualCtx context.Context, actualRequest *Request, target map[string]interface{}) error {
				suite.Equal(suite.expectedCtx, actualCtx)
				suite.True(expectedRequest == actualRequest)
				target["first"] = "true"
				return nil
			}),
			ClaimBuilderFunc(func(actualCtx context.Context, actualRequest *Request, target map[string]interface{}) error {
				suite.Equal(suite.expectedCtx, actualCtx)
				suite.True(expectedRequest == actualRequest)
				return suite.expectedErr
			}),
			ClaimBuilderFunc(func(actualCtx context.Context, actualRequest *Request, target map[string]interface{}) error {
				suite.Fail("This claim builder should not have been called")
				return nil
			}),
		}
	)

	suite.Error(
		builder.AddClaims(suite.expectedCtx, expectedRequest, actual),
	)

	suite.Equal(expected, actual)
}

func TestClaimBuilders(t *testing.T) {
	suite.Run(t, new(ClaimBuildersTestSuite))
}

type RequestClaimBuilderTestSuite struct {
	suite.Suite
}

func (suite *RequestClaimBuilderTestSuite) Test() {
	cases := []struct {
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

	for i, testCase := range cases {
		suite.Run(strconv.Itoa(i), func() {
			actual := make(map[string]interface{})
			suite.NoError(
				requestClaimBuilder{}.AddClaims(context.Background(), testCase.request, actual),
			)

			suite.Equal(testCase.expected, actual)
		})
	}
}

func TestRequestClaimBuilder(t *testing.T) {
	suite.Run(t, new(RequestClaimBuilderTestSuite))
}

type StaticClaimBuilderTestSuite struct {
	suite.Suite
}

func (suite *StaticClaimBuilderTestSuite) Test() {
	cases := []struct {
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

	for i, testCase := range cases {
		suite.Run(strconv.Itoa(i), func() {
			actual := make(map[string]interface{})
			suite.NoError(
				testCase.builder.AddClaims(context.Background(), new(Request), actual),
			)

			suite.Equal(testCase.expected, actual)
		})
	}
}

func TestStaticClaimBuilder(t *testing.T) {
	suite.Run(t, new(StaticClaimBuilderTestSuite))
}

type TimeClaimBuilderTestSuite struct {
	suite.Suite
	expectedNow time.Time
}

var _ suite.SetupAllSuite = (*TimeClaimBuilderTestSuite)(nil)

func (suite *TimeClaimBuilderTestSuite) SetupSuite() {
	suite.expectedNow = time.Now()
}

func (suite *TimeClaimBuilderTestSuite) now() time.Time {
	return suite.expectedNow
}

func (suite *TimeClaimBuilderTestSuite) TestX() {
	cases := []struct {
		builder  timeClaimBuilder
		expected map[string]interface{}
	}{
		{
			builder: timeClaimBuilder{
				now:              suite.now,
				disableNotBefore: true,
			},
			expected: map[string]interface{}{
				"iat": suite.expectedNow.UTC().Unix(),
			},
		},
		{
			builder: timeClaimBuilder{
				now: suite.now,
			},
			expected: map[string]interface{}{
				"iat": suite.expectedNow.UTC().Unix(),
				"nbf": suite.expectedNow.UTC().Unix(),
			},
		},
		{
			builder: timeClaimBuilder{
				now:            suite.now,
				duration:       24 * time.Hour,
				notBeforeDelta: 5 * time.Minute,
			},
			expected: map[string]interface{}{
				"iat": suite.expectedNow.UTC().Unix(),
				"nbf": suite.expectedNow.UTC().Add(5 * time.Minute).Unix(),
				"exp": suite.expectedNow.UTC().Add(24 * time.Hour).Unix(),
			},
		},
		{
			builder: timeClaimBuilder{
				now:              suite.now,
				duration:         30 * time.Minute,
				disableNotBefore: true,
			},
			expected: map[string]interface{}{
				"iat": suite.expectedNow.UTC().Unix(),
				"exp": suite.expectedNow.UTC().Add(30 * time.Minute).Unix(),
			},
		},
	}

	for i, testCase := range cases {
		suite.Run(strconv.Itoa(i), func() {
			actual := make(map[string]interface{})
			suite.NoError(
				testCase.builder.AddClaims(context.Background(), new(Request), actual),
			)

			suite.Equal(testCase.expected, actual)
		})
	}
}

func TestTimeClaimBuilder(t *testing.T) {
	suite.Run(t, new(TimeClaimBuilderTestSuite))
}

type NonceClaimBuilderTestSuite struct {
	suite.Suite
	noncer      *randomtest.Noncer
	builder     nonceClaimBuilder
	expectedErr error
}

var _ suite.SetupTestSuite = (*NonceClaimBuilderTestSuite)(nil)
var _ suite.TearDownTestSuite = (*NonceClaimBuilderTestSuite)(nil)

func (suite *NonceClaimBuilderTestSuite) SetupTest() {
	suite.noncer = new(randomtest.Noncer)
	suite.builder = nonceClaimBuilder{n: suite.noncer}
	suite.expectedErr = errors.New("expected")
}

func (suite *NonceClaimBuilderTestSuite) TearDownTest() {
	suite.noncer.AssertExpectations(suite.T())
}

func (suite *NonceClaimBuilderTestSuite) TestSuccess() {
	actual := make(map[string]interface{})
	suite.noncer.ExpectNonce().Return("test", error(nil)).Once()
	suite.NoError(
		suite.builder.AddClaims(context.Background(), new(Request), actual),
	)

	suite.Equal(
		map[string]interface{}{"jti": "test"},
		actual,
	)
}

func (suite *NonceClaimBuilderTestSuite) TestError() {
	actual := make(map[string]interface{})
	suite.noncer.ExpectNonce().Return("", suite.expectedErr).Once()
	suite.Equal(
		suite.expectedErr,
		suite.builder.AddClaims(context.Background(), new(Request), actual),
	)

	suite.Empty(actual)
}

func TestNonceClaimBuilder(t *testing.T) {
	suite.Run(t, new(NonceClaimBuilderTestSuite))
}

type RemoteClaimBuilderTestSuite struct {
	suite.Suite
	server         *httptest.Server
	goodURL        string
	badURL         string
	expectedMethod string
}

var _ suite.SetupAllSuite = (*RemoteClaimBuilderTestSuite)(nil)
var _ suite.TearDownAllSuite = (*RemoteClaimBuilderTestSuite)(nil)

func (suite *RemoteClaimBuilderTestSuite) SetupSuite() {
	mux := http.NewServeMux()
	mux.HandleFunc("/good", suite.goodHandler)
	mux.HandleFunc("/bad", suite.badHandler)
	suite.server = httptest.NewServer(mux)
	suite.goodURL = suite.server.URL + "/good"
	suite.badURL = suite.server.URL + "/bad"
}

func (suite *RemoteClaimBuilderTestSuite) TearDownSuite() {
	suite.server.Close()
}

func (suite *RemoteClaimBuilderTestSuite) TearDownTest() {
	suite.expectedMethod = ""
}

func (suite *RemoteClaimBuilderTestSuite) goodHandler(response http.ResponseWriter, request *http.Request) {
	suite.Equal("application/json", request.Header.Get("Content-Type"))
	expectedMethod := suite.expectedMethod
	if len(expectedMethod) == 0 {
		expectedMethod = http.MethodPost
	}

	suite.Equal(expectedMethod, request.Method)

	b, err := ioutil.ReadAll(request.Body)
	suite.NoError(err)

	var input map[string]interface{}
	suite.NoError(json.Unmarshal(b, &input))
	if input == nil {
		input = make(map[string]interface{})
	}

	input["custom"] = "value"

	response.Header().Set("Content-Type", "application/json")
	b, err = json.Marshal(input)
	suite.NoError(err)
	suite.NotEmpty(b)
	response.Write(b)
}

func (suite *RemoteClaimBuilderTestSuite) badHandler(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("this is not JSON"))
}

func (suite *RemoteClaimBuilderTestSuite) TestAddClaims() {
	cases := []struct {
		method   string
		client   xhttpclient.Interface
		metadata map[string]interface{}
		request  *Request
		expected map[string]interface{}
	}{
		{
			request:  new(Request),
			expected: map[string]interface{}{"custom": "value"},
		},
		{
			request:  &Request{Metadata: map[string]interface{}{"request": "value"}},
			expected: map[string]interface{}{"request": "value", "custom": "value"},
		},
		{
			method:   http.MethodPut,
			client:   new(http.Client),
			metadata: map[string]interface{}{"external": "value"},
			request:  new(Request),
			expected: map[string]interface{}{"external": "value", "custom": "value"},
		},
		{
			method:   http.MethodPatch,
			client:   new(http.Client),
			metadata: map[string]interface{}{"external": "value"},
			request:  &Request{Metadata: map[string]interface{}{"request": "value"}},
			expected: map[string]interface{}{"external": "value", "request": "value", "custom": "value"},
		},
	}

	for i, testCase := range cases {
		suite.Run(strconv.Itoa(i), func() {
			var (
				actual = make(map[string]interface{})

				remoteClaims = &RemoteClaims{
					URL:    suite.goodURL,
					Method: testCase.method,
				}

				builder, err = newRemoteClaimBuilder(
					testCase.client,
					testCase.metadata,
					remoteClaims,
				)
			)

			suite.Require().NoError(err)
			suite.Require().NotNil(builder)
			suite.expectedMethod = testCase.method

			suite.Require().NoError(
				builder.AddClaims(context.Background(), testCase.request, actual),
			)

			suite.Equal(testCase.expected, actual)
		})
	}
}

func (suite *RemoteClaimBuilderTestSuite) TestError() {
	builder, err := newRemoteClaimBuilder(nil, nil, &RemoteClaims{URL: suite.badURL})
	suite.Require().NoError(err)
	suite.Require().NotNil(builder)

	suite.Error(
		builder.AddClaims(context.Background(), new(Request), make(map[string]interface{})),
	)
}

func (suite *RemoteClaimBuilderTestSuite) TestNoURL() {
	builder, err := newRemoteClaimBuilder(new(http.Client), nil, new(RemoteClaims))
	suite.Nil(builder)
	suite.Error(err)
}

func (suite *RemoteClaimBuilderTestSuite) TestBadURL() {
	var (
		remoteClaims = &RemoteClaims{
			URL: "this is not valid (%$&@!()&*()*%",
		}

		builder, err = newRemoteClaimBuilder(new(http.Client), nil, remoteClaims)
	)

	suite.Nil(builder)
	suite.Error(err)
}

func TestRemoteClaimBuilder(t *testing.T) {
	suite.Run(t, new(RemoteClaimBuilderTestSuite))
}

type NewClaimBuildersTestSuite struct {
	suite.Suite
	server      *httptest.Server
	noncer      *randomtest.Noncer
	expectedNow time.Time
}

var _ suite.SetupTestSuite = (*NewClaimBuildersTestSuite)(nil)
var _ suite.SetupAllSuite = (*NewClaimBuildersTestSuite)(nil)
var _ suite.TearDownTestSuite = (*NewClaimBuildersTestSuite)(nil)
var _ suite.TearDownAllSuite = (*NewClaimBuildersTestSuite)(nil)

func (suite *NewClaimBuildersTestSuite) SetupSuite() {
	suite.server = httptest.NewServer(
		http.HandlerFunc(suite.handleRemoteClaims),
	)
}

func (suite *NewClaimBuildersTestSuite) SetupTest() {
	suite.noncer = new(randomtest.Noncer)
	suite.expectedNow = time.Now()
}

func (suite *NewClaimBuildersTestSuite) TearDownSuite() {
	suite.server.Close()
}

func (suite *NewClaimBuildersTestSuite) TearDownTest() {
	suite.noncer.AssertExpectations(suite.T())
}

func (suite *NewClaimBuildersTestSuite) now() time.Time {
	return suite.expectedNow
}

func (suite *NewClaimBuildersTestSuite) replaceNow(cb ClaimBuilders) {
	for _, b := range cb {
		if tb, ok := b.(*timeClaimBuilder); ok {
			tb.now = suite.now
		}
	}
}

func (suite *NewClaimBuildersTestSuite) rawMessage(v interface{}) json.RawMessage {
	raw, err := json.Marshal(v)
	suite.Require().NoError(err)
	return json.RawMessage(raw)
}

func (suite *NewClaimBuildersTestSuite) handleRemoteClaims(response http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	suite.Require().NoError(err)

	var metadata map[string]interface{}
	suite.Require().NoError(json.Unmarshal(body, &metadata))
	suite.Equal(map[string]interface{}{"extra": "extra stuff"}, metadata)

	response.Header().Set("Content-Type", "application/json")
	response.Write([]byte(`{"remote": "value"}`))
}

func (suite *NewClaimBuildersTestSuite) TestMinimum() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
	})

	suite.Require().NoError(err)
	suite.Require().NotEmpty(builder)

	actual := make(map[string]interface{})
	suite.NoError(
		builder.AddClaims(context.Background(), &Request{Claims: map[string]interface{}{"request": 123}}, actual),
	)

	suite.Equal(
		map[string]interface{}{"request": 123},
		actual,
	)
}

func (suite *NewClaimBuildersTestSuite) TestMissingKey() {
	suite.Run("Claims", suite.testClaimsMissingKey)
	suite.Run("Metadata", suite.testMetadataMissingKey)
}

func (suite *NewClaimBuildersTestSuite) testClaimsMissingKey() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Claims: []Value{
			{}, // the value should have something configured
		},
	})

	suite.Nil(builder)
	suite.Error(err)
}

func (suite *NewClaimBuildersTestSuite) testMetadataMissingKey() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Metadata: []Value{
			{}, // the value should have something configured
		},
		Remote: &RemoteClaims{},
	})

	suite.Nil(builder)
	suite.Error(err)
}

func (suite *NewClaimBuildersTestSuite) TestMissingValue() {
	suite.Run("Claims", suite.testClaimsMissingValue)
	suite.Run("Metadata", suite.testMetadataMissingValue)
}

func (suite *NewClaimBuildersTestSuite) testClaimsMissingValue() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Claims: []Value{
			{
				Key: "test",
				// either JSON or Value should be set
			},
		},
	})

	suite.Nil(builder)
	suite.Error(err)
}

func (suite *NewClaimBuildersTestSuite) testMetadataMissingValue() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Metadata: []Value{
			{
				Key: "test",
				// either JSON or Value should be set
			},
		},
		Remote: &RemoteClaims{},
	})

	suite.Nil(builder)
	suite.Error(err)
}

func (suite *NewClaimBuildersTestSuite) TestBadJSONValue() {
	suite.Run("Claims", suite.testClaimsBadJSONValue)
	suite.Run("Metadata", suite.testMetadataBadJSONValue)
}

func (suite *NewClaimBuildersTestSuite) testClaimsBadJSONValue() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Claims: []Value{
			{
				Key:  "test",
				JSON: `{"this isn't valid JSON`,
			},
		},
	})

	suite.Nil(builder)
	suite.Error(err)
}

func (suite *NewClaimBuildersTestSuite) testMetadataBadJSONValue() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Metadata: []Value{
			{
				Key:  "test",
				JSON: `{"this isn't valid JSON`,
			},
		},
		Remote: &RemoteClaims{},
	})

	suite.Nil(builder)
	suite.Error(err)
}

func (suite *NewClaimBuildersTestSuite) TestStatic() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Claims: []Value{
			{
				Key:   "static1",
				Value: -72.5,
			},
			{
				Key:   "static2",
				Value: []string{"a", "b"},
			},
			{
				Key:    "http1",
				Header: "X-Ignore-Me",
			},
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotEmpty(builder)

	actual := make(map[string]interface{})
	suite.NoError(
		builder.AddClaims(context.Background(), &Request{Claims: map[string]interface{}{"request": 123}}, actual),
	)

	suite.Equal(
		map[string]interface{}{
			"static1": suite.rawMessage(-72.5),
			"static2": suite.rawMessage([]string{"a", "b"}),
			"request": 123,
		},
		actual,
	)
}

func (suite *NewClaimBuildersTestSuite) TestNoRemote() {
	builder, err := NewClaimBuilders(suite.noncer, nil, Options{
		Nonce:          true,
		Duration:       24 * time.Hour,
		NotBeforeDelta: 15 * time.Second,
		Claims: []Value{
			{
				Key:   "static1",
				Value: -72.5,
			},
			{
				Key:   "static2",
				Value: []string{"a", "b"},
			},
			{
				Key:    "http1",
				Header: "X-Ignore-Me",
			},
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotEmpty(builder)

	suite.replaceNow(builder)
	suite.noncer.ExpectNonce().Return("test", error(nil)).Once()

	actual := make(map[string]interface{})
	suite.NoError(
		builder.AddClaims(context.Background(), &Request{Claims: map[string]interface{}{"request": 123}}, actual),
	)

	suite.Equal(
		map[string]interface{}{
			"static1": suite.rawMessage(-72.5),
			"static2": suite.rawMessage([]string{"a", "b"}),
			"request": 123,
			"jti":     "test",
			"iat":     suite.expectedNow.UTC().Unix(),
			"nbf":     suite.expectedNow.Add(15 * time.Second).UTC().Unix(),
			"exp":     suite.expectedNow.Add(24 * time.Hour).UTC().Unix(),
		},
		actual,
	)
}

func (suite *NewClaimBuildersTestSuite) TestBadRemote() {
	_, err := NewClaimBuilders(nil, nil, Options{
		Nonce:          true,
		Duration:       24 * time.Hour,
		NotBeforeDelta: 15 * time.Second,
		Metadata: []Value{
			{
				Key:   "extra1",
				Value: "extra stuff",
			},
			{
				Key:       "http2",
				Parameter: "foo",
			},
		},
		Remote: &RemoteClaims{}, // invalid: missing a URL
	})

	suite.Error(err)
}

func (suite *NewClaimBuildersTestSuite) TestFull() {
	var (
		options = Options{
			Nonce:          true,
			Duration:       24 * time.Hour,
			NotBeforeDelta: 15 * time.Second,
			Claims: []Value{
				{
					Key:   "static1",
					Value: -72.5,
				},
				{
					Key:   "static2",
					Value: []string{"a", "b"},
				},
				{
					Key:    "http1",
					Header: "X-Ignore-Me",
				},
			},
			Metadata: []Value{
				{
					Key:   "extra",
					Value: "extra stuff",
				},
				{
					Key:       "http2",
					Parameter: "foo",
				},
			},
			Remote: &RemoteClaims{
				URL: suite.server.URL,
			},
		}
	)

	builder, err := NewClaimBuilders(suite.noncer, nil, options)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(builder)

	suite.replaceNow(builder)
	suite.noncer.ExpectNonce().Return("test", error(nil)).Once()

	actual := make(map[string]interface{})
	suite.NoError(
		builder.AddClaims(context.Background(), &Request{Claims: map[string]interface{}{"request": 123}}, actual),
	)

	suite.Equal(
		map[string]interface{}{
			"static1": suite.rawMessage(-72.5),
			"static2": suite.rawMessage([]string{"a", "b"}),
			"request": 123,
			"remote":  "value",
			"jti":     "test",
			"iat":     suite.expectedNow.UTC().Unix(),
			"nbf":     suite.expectedNow.Add(15 * time.Second).UTC().Unix(),
			"exp":     suite.expectedNow.Add(24 * time.Hour).UTC().Unix(),
		},
		actual,
	)
}

func TestNewClaimBuilders(t *testing.T) {
	suite.Run(t, new(NewClaimBuildersTestSuite))
}
