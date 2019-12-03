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
- [How to Install](#how-to-install)
- [Usage](#usage)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Details
Themis provides a flexible strategy to issue JWT tokens to devices that need to connect to the XMiDT cluster. 


### Running modes driven by configuration
One of the areas of great flexibility in themis is the configurable origin of claims for outgoing tokens. The claims can be statistically provided in config and/or dynamically by specifying a remote server that serves them.

Static Claims:
```
./themis -f themis.yaml

curl -X GET \
  http://localhost:6501/issue \
  -H 'X-Midt-Mac-Address: mac:1122334455'
```

Dynamic claims:
```
# configuration file specifies static claims as well as a remote server for dynamic ones
# Note: CPE stands for Customer Premise Equipment (term used at Comcast to refer to 
# customer devices) and represents the running mode in which Themis issues JWT tokens
# to devices
./themis -f cpe_themis.yaml 


# Start the server that will serve the claims mentioned above
# Note: RBL stands for Remote Business Logic which represents a mode in which themis 
# simply serves claims 
./themis -f rbl_themis

# make the request and observe the additional claims from the "remote" server

``` 

## How to Install

### Installation
- [Docker](https://www.docker.com/) (duh)
  - `brew install docker`

</br>

### Running
#### Build the docker image
```bash
docker build -t themis:local .
```
This `build.sh` script will build the binary and docker image

## Usage
Once everything is up and running you can start sending requests. Below are a few examples.
TODO: Add examples

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
