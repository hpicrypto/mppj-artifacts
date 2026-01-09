# Multi-Party Private Join – Artifacts

This repository hosts the artifact for the paper "Multi-Party Private Join" by Anja Lehmann,
Christian Mouchet and Andrey Sidorenko, PETS 2026. This artifact is a set of scripts and
utilities to run the performance benchmarks. The main MPPJ protocol implementation is
imported from [https://github.com/hpicrypto/mppj].

For a complete procedure to run the experiments and reproduce the paper's results, see
[ARTIFACT-APPENDIX.md]. For more information on the protocol's implementation and its
most up-to-date version, see [https://github.com/hpicrypto/mppj]. 

## Directory structure

- `mppj-exps`: an expperiment runner for the MPPJ protocol.
- `mpspdz-exps`: an experiment runner for the generic MPC baseline (MP-SPDZ-based).
- `setup`: some additional setup-utilities for experiments over networked machines.

## Security

This repository contains a prototype implementation of the MPPJ protocol. This is for
academic research purposes and should not be considered production-ready.