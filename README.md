# themis

[![Build Status](https://travis-ci.com/xmidt-org/themis.svg?branch=master)](https://travis-ci.com/xmidt-org/themis) 
[![codecov.io](http://codecov.io/github/xmidt-org/themis/coverage.svg?branch=master)](http://codecov.io/github/xmidt-org/themis?branch=master)
[![Code Climate](https://codeclimate.com/github/xmidt-org/themis/badges/gpa.svg)](https://codeclimate.com/github/xmidt-org/themis)
[![Issue Count](https://codeclimate.com/github/xmidt-org/themis/badges/issue_count.svg)](https://codeclimate.com/github/xmidt-org/themis)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/themis)](https://goreportcard.com/report/github.com/xmidt-org/themis)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/themis/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/xmidt-org/themis?include_prereleases)](CHANGELOG.md)

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
docker build -t themis:local .
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
