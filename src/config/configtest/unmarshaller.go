package configtest

import "github.com/stretchr/testify/mock"

// Unmarshaller is a mocked config.Unmarshaller and config.KeyUnmarshaller
type Unmarshaller struct {
	mock.Mock
}

func (u *Unmarshaller) Unmarshal(v interface{}) error {
	return u.Called(v).Error(0)
}

func (u *Unmarshaller) ExpectUnmarshal(v interface{}) *mock.Call {
	return u.On("Unmarshal", v)
}

func (u *Unmarshaller) UnmarshalKey(key string, v interface{}) error {
	return u.Called(key, v).Error(0)
}

func (u *Unmarshaller) ExpectUnmarshalKey(key string, v interface{}) *mock.Call {
	return u.On("UnmarshalKey", key, v)
}
