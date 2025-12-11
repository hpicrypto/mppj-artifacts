# MPPJ-Go: a Go Implementation of the DH-MPPJ Protocol

This package implement the FD-MPPJ protocol, as proposed in the paper "Multi-party Private
Joins" by by Anja Lehmann, Christian Mouchet and Andrey Sidorenko, PETS 2026. It implement
this protocol over the P-256 elliptic curve. 

## Package Structure

- `party_datasource.go`: the source-related operations.
- `party_helper.go`: the helper-related operations.
- `party_receiver.go`: the receiver-related operations.
- `group.go` a group abstraction for ElGamal.
- `encryption.go` the PKE / SE functionality
- `prf.go` the Hash-DH OPRF (for ElGamal PKE)
- `table.go` some basic types (plaintext table, joined table) and functions for tables
- `mppj_test.go` some end-to-end tests.
- `benchmark_test.go` some micro-benchmarks for individual operations.
- `api` a gRPC-based service for the helper (server) and source/receiver (clients).
- `cmd` the executables (main packages) for the sources/helper/receiver.

## Current Limitations

- The number of sources is limited to 256, as it encode the origin table over a single byte.
- The values are also assumed to be smaller than 30 bytes, for encoding them as group
  elements in a decodable way.
- The large-values extension proposed of the paper is not yet implemented.

## Security

This repository contains a prototype implementation of the MPPJ protocol. This is for
academic research purposes and should not be considered production-ready. Notably, the
code was not externally audited and includes several non-constant-time algorithms.
