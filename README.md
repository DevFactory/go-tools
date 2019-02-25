# Go tools
[![GoDoc](https://godoc.org/github.com/DevFactory/go-tools?status.svg)](https://godoc.org/github.com/DevFactory/go-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/DevFactory/go-tools)](https://goreportcard.com/report/github.com/DevFactory/go-tools)
[![Build Status](https://travis-ci.com/DevFactory/go-tools.svg?branch=master)](https://travis-ci.com/DevFactory/go-tools)

This repository includes various tools and extensions that seemed to be missing in the standard go library and well known open source projects.
Dependencies are managed with `dep`.

## Content
* `pkg/extensions`
  * `collections` - extensions for operations on collections
  * `net` - utilities to convert and operate on structures returned by the standard go's `net` package
  * `os` - improvements to go's `os` package, currently mostly about loading env vars with defaults
  * `strings` - to operate on slices of strings
  * `time` - a `refresher` utility to periodically execute a callback function
* `pkg/linux/command` - execute linux commands and capture full output info: stdOut, stdErr and exitCode
* `nettools` - various utilities that wrap basic Linux network configuration commands, including
  * `conntrack` - currently only supports removing entries from conntrack
  * `interfaces` - allows for abstract implementation of `net` package interface info
  * `iproute` - management of `ip route` route table entries and `ip rule` routing rules
  * `ipset` - management of `ipset`: sets and set entries
  * `iptables` - management of `iptables` rules based on unique rule comments

## Getting started - contributing
1. Install [dep](https://github.com/golang/dep#installation)
1. Clone this repository to `$GOPATH/src/github.com/DevFactory/go-tools`
1. Install dependencies with `dep`: `dep ensure -v --vendor-only`
1. Everything ready, you can run unit tests: `go test  -cover ./pkg/...`
