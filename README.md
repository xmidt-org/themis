# Themis

[![Build Status](https://travis-ci.com/xmidt-org/themis.svg?branch=master)](https://travis-ci.com/xmidt-org/themis)
[![codecov.io](http://codecov.io/github/xmidt-org/themis/coverage.svg?branch=master)](http://codecov.io/github/xmidt-org/themis?branch=master)
[![Code Climate](https://codeclimate.com/github/xmidt-org/themis/badges/gpa.svg)](https://codeclimate.com/github/xmidt-org/themis)
[![Issue Count](https://codeclimate.com/github/xmidt-org/themis/badges/issue_count.svg)](https://codeclimate.com/github/xmidt-org/themis)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/themis)](https://goreportcard.com/report/github.com/xmidt-org/themis)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/themis/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/xmidt-org/themis?include_prereleases)](CHANGELOG.md)

## Summary

A JWT token issuer for devices that connect to the XMiDT cluster.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Details](#details)
- [Build](#build)
- [Deploy](#deploy)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Details
Themis provides a flexible strategy to issue JWT tokens to devices that need to connect to the XMiDT cluster. 

### Endpoints
There are three main endpoints (directly mapped to servers `key`, `issuer` and `claims` in configuration) this service provides:

- GET `/keys/{KID}`

This endpoint allows fetching the public portion of the key that themis uses to sign JWT tokens. For example, [Talaria](https://github.com/xmidt-org/talaria) can use this endpoint to verify the signature of tokens which devices present when they attempt to connect to XMiDT.

Configuration for this endpoint is required when the `issue` endpoint is configured and vice versa.

- POST `/issue`

This is the main and most compute intensive Themis endpoint as it creates JWT tokens based on configuration. 

- GET `/claims`

Configuring this endpoint is required if no configuration is provided for the previous two.


### JWT Claims Configuration
Claims can be configured through the `token.claims`, `partnerID` and `remote` configuration elements. The claim values themselves can come from multiple sources.

#### Fixed values in configuration
```
token:
  ...

  claims:
    capabilities
      value:
        - capability0
        - capability1

```
The above config would create the claim: 
``` 
"capabilities":  ["capability0", "capability1"]
```
#### HTTP Header or Parameter 
```
token:  
  ...

  claims:
    mac:
      header: X-Midt-Mac-Address
      parameter: mac
```
The value of the `mac` claim would come from the specified header or parameter name of the request to the `/issue` endpoint.

#### PartnerID
Although it is configured separately, it behaves very similarly to the previous source type.

```
partnerID:
  claim: partner-id
  metadata: pid # only needed when a remote claims server needs this value
  header: X-Midt-Partner-ID
  parameter: pid
  default: comcast
```

#### Remote claims

```
remote:
  method: "POST"
  url: "http://remote-claims-server.example.com/claims"
```
For more informatiom on how to configure Themis to run as your remote claims server, read the next section on Remote Server Claims Configuration.


### Remote Server Claims Configuration
### Using Themis as the remote claims server
You can do this by configuring only the `claims` server in your configuration file. 
Claims are configured exactly the same as explained above.

#### Sending data to remote server
Suppose the remote claims server needs the ID of the device requesting a token in the form of an HTTP Header named `X-Midt-Device-Id`. The `token.metadata` configuration element allows you to specify which values are sent to the remote claims server.

```
token:
  ...
  metadata:
    ID:
      header: X-Midt-Device-Id
```

## Build
There is a single binary for themis and its execution is fully driven by configuration.

### Makefile

The Makefile has the following options you may find helpful:
* `make build`: builds the Themis binary
* `make docker`: builds a docker image for themis
* `make local-docker`: builds a docker image for themis with the `local` version tag
* `make test`: runs unit tests with coverage for Themis 
* `make clean`: deletes previously-built binaries and object files

## Deploy
At the simplest form, run the binary with the flag specifying the configuration file
```
./themis -f themis.yaml
``` 

### Docker
We recommend using docker for local development.

```
# Build docker image for themis
# themis.yaml specifies the static claims which will be returned in the JWT
make local-docker

# Run container service
docker run -p 6501:6501 themis:local

# Request a JWT token
curl http://localhost:6701/issue -H 'X-Midt-Mac-Address: mac:1122334455'
```

Explore JWT contents at [jwt.io](https://jwt.io/)

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).