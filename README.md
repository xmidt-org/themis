# xmidt-issuer

[![Build Status](https://travis-ci.com/Comcast/xmidt-issuer.svg?branch=master)](https://travis-ci.com/Comcast/xmidt-issuer) 
[![codecov.io](http://codecov.io/github/Comcast/xmidt-issuer/coverage.svg?branch=master)](http://codecov.io/github/Comcast/xmidt-issuer?branch=master)
[![Code Climate](https://codeclimate.com/github/Comcast/xmidt-issuer/badges/gpa.svg)](https://codeclimate.com/github/Comcast/xmidt-issuer)
[![Issue Count](https://codeclimate.com/github/Comcast/xmidt-issuer/badges/issue_count.svg)](https://codeclimate.com/github/Comcast/xmidt-issuer)
[![Go Report Card](https://goreportcard.com/badge/github.com/Comcast/xmidt-issuer)](https://goreportcard.com/report/github.com/Comcast/xmidt-issuer)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/Comcast/xmidt-issuer/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/Comcast/xmidt-issuer.svg)](CHANGELOG.md)

The Xmidt server that provides a framework for issueing JWT tokens to connecting devices.

# How to Install

## Dockerized Caduceus
Docker containers make life super easy.

### Installation
- [Docker](https://www.docker.com/) (duh)
  - `brew install docker`

</br>

### Running
#### Build the docker image
```bash
docker build -t xmidt-issuer:local .
```
This `build.sh` script will build the binary and docker image

#### Run the image

## Usage
Once everything is up and running you can start sending requests. Below are a few examples.

#### Get Health
```bash
curl http://localhost:{TBD}/health
```

#### GET some [pprof](https://golang.org/pkg/net/http/pprof/) stats
```bash
curl http://localhost:{TBD}/debug/pprof/mutex
```
