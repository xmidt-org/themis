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
By participating, you agree to this code.

## Details
Themis provides a flexible strategy to issue JWT tokens to devices that need to connect to the XMiDT cluster. 

### Endpoints
There are three main endpoints (directly mapped to servers `key`, `issuer` and `claims` in configuration) this service provides

- GET `/keys/{KID}`

This endpoint allows fetching the public portion of the key that themis uses to sign JWT tokens. For example, [Talaria](https://github.com/xmidt-org/talaria) can use this endpoint to verify the signature of tokens which devices present when they attempt to connect to the XMiDT cloud.

Configuration for this endpoint is required when the `issue` endpoint is configured and vice versa.

- POST `/issue`

This is the main and most compute intensive endpoint as it creates JWT tokens based on configuration. XMiDT can be configured such that it only accepts devices that have valid JWT tokens.

- GET `/claims`

Configuring this endpoint is required if no configuration is provided for the previous two.


## Build
There is a single binary for themis and its execution is fully driven by configuration.

### Makefile

The Makefile has the following options you may find helpful:
* `make build`: builds the Themis binary
* `make docker`: builds a docker image for themis
* `make cpe-docker`: builds a docker image for themis in CPE mode
* `make rbl-docker`: builds a docker image for themis in RBL mode 

* `make test`: runs unit tests with coverage for Themis 
* `make clean`: deletes previously-built binaries and object files

## Deploy
At the simplest form, run the binary with the flag specifying the configuration file
```
./themis -f configuration.yaml
``` 

### Docker
We recommend using docker for local development.

There are two current intended use cases for themis which determine the deployment path.

1) JWT Token claims are provided to Themis through configuration 
```
# Build docker image for themis
# /deploy/config/themis.yaml specifies the static claims 
docker build -t themis:local -f deploy/docker/Dockerfile

# Run container service
docker run -p 6701:6701 themis:local

# Request a JWT token 
curl http://localhost:6701/issue -H 'X-Midt-Mac-Address: mac:1122334455'
```

2) JWT Token claims are provided to themis both through configuration AND a configurable remote server 

```
# Build docker image for cpe-themis
# /deploy/docker/modes/cpe/cpe_themis.yaml specifies static claims as well as a 
# remote server for dynamic ones
# Note: CPE stands for Customer Premise Equipment (term used at Comcast to refer to 
# customer devices) and represents the running mode in which Themis issues JWT tokens
# to devices
docker build -t cpe_themis:local -f deploy/docker/modes/cpe/Dockerfile .


# Build docker image for rbl-themis 
# This will allow running the remote claims server used by cpe-themis
# /deploy/config/rbl_themis.yaml specifies static claims that the service will serve
# Note: RBL stands for Remote Business Logic which represents a mode in which themis 
# simply serves claims 
docker build -t rbl_themis:local -f deploy/docker/modes/rbl/Dockerfile .

# Start a cluster of cpe and rbl themises 
docker-compose -f deploy/docker/modes/docker-compose.yaml up

# Request a JWT token and observe the additional claims from the "remote" server
# on JWT
curl http://localhost:6501/issue -H 'X-Midt-Mac-Address: mac:1122334455'
```
## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
