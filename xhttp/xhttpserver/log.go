package xhttpserver

const (
	addressKey = "address"
	serverKey  = "server"
)

// AddressKey is the logging key for the server's bind address
func AddressKey() interface{} {
	return addressKey
}

// ServerKey is the logging key for the server's name
func ServerKey() interface{} {
	return serverKey
}
