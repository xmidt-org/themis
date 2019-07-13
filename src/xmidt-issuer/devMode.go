package main

const (
	// devMode is the YAML configuration parsed by Viper when the server is run with the dev mode flag
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

responseHeaders:
  X-Midt-Server: issuer
  X-Midt-Version: development

token:
  alg: RS256
  nonce: true
  notBeforeDelta: -15s
  duration: 24h
  requestClaims:
    mac:
      header: X-Midt-Mac-Address
      parameter: mac
      required: true
    serial:
      header: X-Midt-Serial-Number
      parameter: serial
      required: true
    uuid:
      header: X-Midt-Uuid
      parameter: uuid
      required: true
  claims:
    iss: development
    partner-id: comcast
    trust: 1000
    sub: "client:supplied"
    aud: XMiDT
    capabilities:
      -
        x1:issuer:test:.*:all
  key:
    kid: development
    type: rsa
    bits: 1024

log:
  file: stdout
  level: DEBUG
`
)
