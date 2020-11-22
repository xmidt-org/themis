package main

const (
	// devMode is the YAML configuration parsed by Viper when the server is run with the dev mode flag
	devMode = `
servers:
  key:
    address: :8080
    disableHTTPKeepAlives: true
    header:
      X-Midt-Server:
        - issuer
      X-Midt-Version:
        - development

  issuer:
    address: :8081
    disableHTTPKeepAlives: true
    header:
      X-Midt-Server:
        - issuer
      X-Midt-Version:
        - development

  claims:
    address: :8082
    disableHTTPKeepAlives: true
    header:
      X-Midt-Server:
        - issuer
      X-Midt-Version:
        - development

  metrics:
    address: :8083
    disableHTTPKeepAlives: true

  pprof:
    address: :9999
    disableHTTPKeepAlives: true

  health:
    address: :8084
    disableHTTPKeepAlives: true
    header:
      X-Midt-Server:
        - issuer
      X-Midt-Version:
        - development

health:
  disableLogging: false
  custom:
    server: development

prometheus:
  defaultNamespace: xmidt
  defaultSubsystem: issuer
  constLabels:
    development: "true"

token:
  alg: RS256
  nonce: true
  notBeforeDelta: -15s
  duration: 24h
  claims:
    - key: mac
      header: X-Midt-Mac-Address
      parameter: mac
    - key: serial
      header: X-Midt-Serial-Number
      parameter: serial
    - key: uuid
      header: X-Midt-Uuid
      parameter: uuid
    - key: iss
      value: "development"
    - key: trust
      value: 1000
    - key: sub
      value: "client-supplied"
    - key: aud
      value: "XMiDT"
    - key: capabilities
      value:
        -
          x1:issuer:test:.*:all
    - key: allowedResources
      json: '{
        "allowedPartners": ["comcast"],
        "allowedServiceAccountIds": ["1234", "5678"]
      }'

  metadata:
    - key: mac
      header: X-Midt-Mac-Address
      parameter: mac
    - key: serial
      header: X-Midt-Serial-Number
      parameter: serial
    - key: uuid
      header: X-Midt-Uuid
      parameter: uuid
  partnerID:
    claim: partner-id
    header: X-Midt-Partner-Id
    parameter: pid
    default: comcast
  key:
    kid: development
    type: rsa
    bits: 1024

log:
  file: stdout
  level: INFO
`
)
