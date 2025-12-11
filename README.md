# Multi-Party Private Join – Artifacts

This repository hosts the artifact for the paper "Multi-Party Private Join" by Anja Lehmann,
Christian Mouchet and Andrey Sidorenko, PETS 2026. This artifact has two main components:
1. An implementation of the MPPJ protocol, as a Go package.
2. A set of scripts and utilities to run the performance benchmarks.

For a complete procedure to run the experiments and reproduce the paper's results, see
`ARTIFACT-APPENDIX.md`. 

For more information on the protocol's implementation, see `mppj-go/README.md`.

## Directory structure

- `mppj-go`: the Go package implementation of the MPPJ protocol.
- `mppj-exps`: an expperiment runner for the MPPJ protocol.
- `mpspdz-exps`: an experiment runner for the generic MPC baseline (MP-SPDZ-based).
- `setup`: some additional setup-utilities for experiments over networked machines.

## Security

This repository contains a prototype implementation of the MPPJ protocol. This is for
academic research purposes and should not be considered production-ready.