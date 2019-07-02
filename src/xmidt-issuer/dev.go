package main

const (
	devMode = `---
servers:
  key:
    address: :8080
    logConnectionState: true
    disableHTTPKeepAlives: true

  issuer:
    address: :8081
    logConnectionState: true
    disableHTTPKeepAlives: true

  metrics:
    address: :8082
    logConnectionState: true
    disableHTTPKeepAlives: true

token:
  alg: RS256
  nonce: true
  key:
    kid: development
    type: rsa
    bits: 1024

log:
  file: stdout
  level: DEBUG
`
)
