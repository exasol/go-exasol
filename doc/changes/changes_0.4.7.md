# exasol-driver-go 0.4.7, released 2022-10-??

Code name: Improve documentation and make internal APIs private.

## Summary

In this release we improved the Godoc documentation available at [pkg.go.dev](https://pkg.go.dev/github.com/exasol/exasol-driver-go). We also made internal types private that are used for the communication with the database.

## Features

* #82: Improve Godoc documentation.
## Dependency Updates

### Test Dependency Updates

* Updated `golang.org/x/sync:v0.0.0-20220722155255-886fb9371eb4` to `v0.1.0`
* Updated `github.com/stretchr/testify:v1.8.0` to `v1.8.1`
* Added `github.com/exasol/exasol-test-setup-abstraction-server/go-client:v0.3.0`

### Other Dependency Updates

* Removed `github.com/testcontainers/testcontainers-go:v0.13.0`