# httpgrpc

[![GoDoc](https://godoc.org/github.com/LLKennedy/httpgrpc?status.svg)](https://godoc.org/github.com/LLKennedy/httpgrpc)
[![Build Status](https://travis-ci.org/disintegration/imaging.svg?branch=master)](https://travis-ci.org/LLKennedy/httpgrpc)
![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/LLKennedy/httpgrpc.svg)
[![Coverage Status](https://coveralls.io/repos/github/LLKennedy/httpgrpc/badge.svg?branch=master)](https://coveralls.io/github/LLKennedy/httpgrpc?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/LLKennedy/httpgrpc)](https://goreportcard.com/report/github.com/LLKennedy/httpgrpc)
<!-- [![Maintainability](https://api.codeclimate.com/v1/badges/22d24397a4cccf8471d4/maintainability)](https://codeclimate.com/github/LLKennedy/httpgrpc/maintainability) -->
[![GitHub](https://img.shields.io/github/license/LLKennedy/httpgrpc.svg)](https://github.com/LLKennedy/httpgrpc/blob/master/LICENSE)

Microservice API to convert external HTTP endpoints on a proxy to internally exposed GRPC messages. Allows a generic proxy to talk to services via a standard message while still allowing each service to maintain its API using GRPC and protocol buffers.

There are multiple implemenations that follow this basic intent already (HTTP+JSON reverse proxied to GRPC) but assume your service is directly handling external HTTP traffic, rather than sitting behind load-balanced webservers in a DMZ somewhere separate to your nice safe application server. For example, [here](https://github.com/grpc-ecosystem/grpc-gateway) and [here](https://github.com/weaveworks/common/tree/master/httpgrpc).

This differs, in that your application is only expected to be handling GRPC. The logic used by the reverse proxy to determine where to send the message is up to you, this library simply sets the standard for what must be passed on - an HTTP method, a procedure name and the JSON payload. Your service must provide a procedure with a name that matches the format MethodnameProcedureName, such as PostLogin or GetUserPhoto.

## Installation

`go get "github.com/LLKennedy/httpgrpc"`

## Basic Usage

TODO

## Testing

On windows, the simplest way to test is to use the powershell script.

`./test.ps1`

To emulate the testing which occurs in build pipelines for linux and mac, run the following:

`go test ./... -race`