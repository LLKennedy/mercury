# httpgrpc

[![GoDoc](https://godoc.org/github.com/LLKennedy/httpgrpc?status.svg)](https://godoc.org/github.com/LLKennedy/httpgrpc)
[![Build Status](https://travis-ci.org/disintegration/imaging.svg?branch=master)](https://travis-ci.org/LLKennedy/httpgrpc)
![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/LLKennedy/httpgrpc.svg)
[![Coverage Status](https://coveralls.io/repos/github/LLKennedy/httpgrpc/badge.svg?branch=master)](https://coveralls.io/github/LLKennedy/httpgrpc?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/LLKennedy/httpgrpc)](https://goreportcard.com/report/github.com/LLKennedy/httpgrpc)
<!-- [![Maintainability](https://api.codeclimate.com/v1/badges/22d24397a4cccf8471d4/maintainability)](https://codeclimate.com/github/LLKennedy/httpgrpc/maintainability) -->
[![GitHub](https://img.shields.io/github/license/LLKennedy/httpgrpc.svg)](https://github.com/LLKennedy/httpgrpc/blob/master/LICENSE)

Microservice API to convert external HTTP endpoints on a proxy to internally exposed GRPC messages. Allows a generic proxy to talk to services via a standard message while still allowing each service to maintain its API using GRPC and protocol buffers.

## Installation
`go get "github.com/LLKennedy/httpgrpc"`

## Basic Usage

TODO

## Testing

On windows, the simplest way to test is to use the powershell script.

`./test.ps1`

To emulate the testing which occurs in build pipelines for linux and mac, run the following:

`go test ./... -race`