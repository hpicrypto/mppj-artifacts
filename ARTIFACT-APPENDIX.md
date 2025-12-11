# Artifact Appendix

Paper title: **Multi-Party Private Join**

Requested Badge(s):
  - [x] **Available**
  - [x] **Functional**
  - [x] **Reproduced**

## Description

This artifact complements the paper "Multi-Party Private Join" by Anja Lehmann,
Christian Mouchet and Andrey Sidorenko, PETS 2026. It provides a reference implementation
for the cryptographic protocol, DH-MPPJ, proposed in this work, along with the necessary
sources for reproducing the experiments. There are two experiments:
1. Experiment 1: The main experiment is an end-to-end benchmark of the DH-MPPJ protocol.
 It is presented in Section 5.3.2 of the paper.
2. Experiment 2: A secondary experiment compares the DH-MPPJ protocol to generic MPC. It
 is presented in Section 5.3.1 of the paper.

The repository is structured as follows:

- `mppj-go`: the source code for the DH-MPPJ protocol, packaged as a standalone Go module.
- `mppj-exps`: the scripts for running experimental evaluation of the end-to-end DH-MPPJ
 protocol. This implements Experiment 1 and the DH-MPPJ part of Experiment 2.
- `mpspdz-exps`: the scripts and MP-SPDZ programs for running the generic-MPC baseline.
 This implements the generic-MPC part of Experiment 2.
- `setup`: some additional utilities for setting up experiments over networked machines.

**Experimental Setup Overview**

This artifact provides a build system for the code, which is based on Docker. Then, the
experiments are Python scripts that run the built images. Specifically, the experimental
setup consists of two main types of machines: 

1. An experiment **orchestrator** machine: it runs a Python script which starts containers
 running the protocol. This would usually be the local machine of the experimenter.
2. One or multiple **Docker hosts** machine(s): these are orchestrated by the orchestrator
 and run the actual parties of the protocol as containers. This can also be the local
 machine of the experimenter, or remote Docker hosts, to enable larger-scale and
 more realistic, networked experiments.

In the rest of this document, we describe the requirements for the orchestrator and the
Docker hosts separately, regardless of whether they are the same machine or not. It is
recommended to start with a fully local setup, then switch to remote Docker hosts for
running larger-scale and more realistic experiments.

Each experiment has two modes (given as an argument to the scripts):

1. `full`: runs the full experiment, as reported in the paper.
2. `light`: runs smaller-scale experiments, and with fewer repetitions. 

We detail the specific requirements for each experiment later in this document. As a rule
of thumb, the `light` experiment set is designed to run with a fully local setup on a
modern laptop, in a few minutes. The `full` experiment set requires more
resources (such as a desktop PC or a server), especially if representative performance
results are expected.

### Security/Privacy Issues and Ethical Concerns

This artifact poses no direct security/privacy risk for its user. The experiments' scripts
in their default configuration will only start Docker containers that bind the `localhost`
address. However, running realistic experiments over network machines should only be done
over a (possibly virtual) private network, as the reference implementation does not
authenticate the connections.

All test data used in the experiments is generated synthetically and on-the-fly. 

## Basic Requirements

This section describes the software and hardware requirements for this artifact.

### Hardware Requirements

The **orchestrator machine** does not have special hardware requirements. We assume it is
any modern x86-64 architecture with at least 4 logical CPUs. The orchestrator machine does
not require any non-negligible disk space (it requires the code repository and stores the
experiments' results, all of which are text files).

The **Docker hosts** also do not have special hardware requirements, but
the amount of available resources will influence the experimental measurements. At a
minimum, a Docker host should have at least 4 logical CPUs and 16 GB RAM available. The
Docker host requires some disk space to build and store the Docker images. After building
the two images required by our experiments, `docker system df` reports a disk usage of ~4
GB. Note that this is mostly used by the MP-SPDZ image. When only the MPPJ image is built,
600 MB of disk space suffices.

**Our Setup**

The experimental setup we used for the paper's result consists of two AWS `c5.4xlarge`
servers with 16 vCPU and 32 GiB of RAM. According to `/proc/cpuinfo`, our instances were
running on an Intel Xeon Platinum 8275CL (3 GHz, 16 threads).

For Experiment 1, we run the helper on one server, and all the sources and the receiver on
the other server. The containers rely on the native host networking stack and communicate
via the local network of the AWS zone. We limit each source to 4 concurrent threads. Note
that, in our protocol, the sources and the receiver are never running concurrently. This
leaves all 16 CPUs for the latter.

For Experiment 2, we run all the parties on one instance, over the localhost interface.

### Software Requirements

The **orchestrator machine** requirements are:
- `git` for pulling the artifact code
- `Python3` (tested with v3.14.1), and `python3-venv`

The **Docker host machines** requirements are:
- `Docker` (tested with Engine v28.1.1)
- `git` for pulling the artifact code
- `make` for building the images (this is a soft requirement, the Makefiles are wrappers
  around `docker build`).

**Optional**

We also provide some automated configuration tools for setting up remote Docker hosts
machines from an inventory file (see the "Setting up remote Docker Hosts" section at the
end of this document). These have the following requirements (typically, for the
experimentor's machine): 
- `openssl` for generating a TLS configuration  (tested with version 3.6.0)
- `yq` for parsing the inventory file (https://github.com/mikefarah/yq)
- `ansible` for configuring the Docker hosts.

Although it is not necessary for reproducing the paper's result, the Go toolchain is
required for locally building/running the unit test and micro-benchmarks.

For reference, this artifact was tested under Ubuntu 24.04.03 LTS Kernel 6.8.0, and MacOS
13.7.8. The code was developped and tested with Go version 1.24 and Python 3.12.11. The
reference implementation manages its dependency automatically through the Go toolchain,
see the `mppj-go/go.mod` file for a list of the dependencies.   

### Estimated Time and Storage Consumption

The overall human time depends on the environment (fully local vs remote Docker hosts).
The compute time depends on which experiments are run, on the experiment grid (full vs
light), and on the available computing resources/network speed. We provide an overview
below. 

- Overall human time: 10-20 minutes
   - Environment setup: 5 min (fully local) - 15 min (remote Docker hosts)
   - Experiments configuration execution: 1 min (fully local) - 5 min (remote Docker host)
- Overall compute time: 2-3 hours
   - Build time: 3 min (Experiment 1 only) - 30 min (Experiment 1+2)
   - Full Benchmarks: 50 min (Experiment 1) - 2.5 h (Experiments 1+2)
   - Light Benchmarks: 10 min (Experiment 1) - 5 min (Experiment 1+2)
- Overall disk usage: 500 MB (Experiment 1) - 4 GB (Experiment 1+2)


## Environment

This section describes how to access the artifact and set up the experiment environment.

### Accessibility

The artifact is available at `https://github.com/hpicrypto/mppj-artifacts`.

### Set up the environment

For **all machines** (orchestrator and Docker host if distinct), clone the artifact
repository: 
```bash
git clone git@github.com:hpicrypto/mppj-artifacts.git 
cd mppj-artifacts
```

For **the orchestrator machine** with the aforementioned software requirements fulfilled,
the environment is set up by installing the required Python packages:

```bash
python3 -m venv .venv && source .venv/bin/activate # uses a virtual python env
pip3 install -r setup/requirements.txt # installs the requirements in the venv
```

For the **Docker hosts machine(s)** with the aforementioned software requirements
fulfilled (see the "Setting up remote Docker hosts" section at the end of this document
for an automated remote host setup procedure), the environment is set up by building the
docker images: 

```bash
make all
```
**Note**: building the generic-MPC baseline (based on MP-SPDZ) takes ~30 min and ~4 GB of
disk space, and is only needed for Experiment 2. For running the MPPJ benchmarks only, the
image can be built individually. 
```bash
make mppj-exps
```
This should complete in ~3 minutes and requires ~600 MB of disk space.


### Testing the Environment

From the Docker host machine(s), testing that the build for the DH-MPPJ implementation has
completed successfully can be done by running the parties' programs. 
```bash
docker run --rm mppj source
docker run --rm mppj helper
docker run --rm mppj receiver
```
All commands should return an error message about some IDs not being provided.

Testing that the MP-SPDZ images built correctly can be done by running:
```bash
docker run --rm -it mpspdz:exp ./Scripts/compile-run.py atlas tutorial
```
This should run the MP-SPDZ "tutorial" tutorial.

## Artifact Evaluation

We now describe the evaluation procedure for this artifact and how it relates to the
claims of the paper.

### Main Results and Claims

To the best of our knowledge, there exists no _standalone_ multi-party private join
protocol at the time of this work. Our only points of comparison are more generic
protocols, specifically, generic MPC and _Private-Join-Compute_ (PJC) protocols. From a
high level, our main claims are threefold:

1. Practical runtime: On realistic datasets and system sizes, our DH-MPPJ completes within
 seconds or minutes.
2. Advantage over generic MPC: The DH-MPPJ protocol outperforms a state-of-the-art
 generic-MPC solution.
3. Advantage over PJC: The DH-MPPJ has a comparable or better performance w.r.t to the
 latest PJC protocol *with no compute phase*.

We detail those claims below.

#### Main Result 1: Practical runtime

Our paper claims that the DH-MPPJ protocol scales favorably with the dataset size $m$ and
the number of parties $n$, leading to practical runtimes for realistic scenarios. We
demonstrate this by running our protocol and measuring the end-to-end performance. 

These results are reported in Table 3 of our paper. This is a two-dimensional table
where we vary $n$ from 2 to 6, and $m$ from ~1300 to ~250'000. These values are chosen to
represent various system and dataset sizes, and align with the figures reported in our PJC
baseline (this is only important for Claim 3).

We observe that the DH-MPPJ has a practical runtime, and it could complete a join for the
largest dataset (~250'000 entries per party) and 6 parties in under 4 minutes on our test
setup. The performance scales linearly with the number of database rows and the number of
parties.

#### Main Result 2: Advantage over Generic MPC

Our paper claims that DH-MPPJ outperforms a generic MPC protocol, already for the two-party
case and for modest dataset sizes. We demonstrate this by running a join circuit with the
ATLAS protocol (as implemented in the MP-SPDZ library), and by running DH-MPPJ for a join 
of the same size. Indeed, this represents a very loose baseline, and is considered a 
_sanity check_. This is why we only consider the two-party case (which typically favors
generic MPC protocols) and three different orders of magnitude for the dataset sizes: 10,
100, 1000.

The results are reported in Table 2 of our paper. This table has only a single dimension,
which is the dataset size. We observe that for a very small dataset, the generic MPC
solution slightly outperforms the DH-MPPJ protocol. Yet, this is mainly due to the network
transmission delays not being included in this experimental setup.


#### Main Result 3: Advantage over PJC:

Our paper claims that DH-MPPJ outperforms the state-of-the-art PJC protocol (stripped
of the computation phase). To demonstrate this, we isolated the most recent PJC work that 
(i) supports $n > 2$ parties and (ii) can be stripped of the "compute" part: the
[IDCloak](https://arxiv.org/abs/2506.01072) protocol. Unfortunately, we could not
successfully run the artifacts of this work in our experimental setup (the implementation
currently lacks a complete network stack and assumes localhost execution), and our comparison
points are limited to the figures reported in their paper. Therefore, we aligned our
experiment grid to these figures ($n \in [2,6]$ and $m \in \{1353, 1700, 19735, 45211, 253680\}$,
$m$ corresponds to the sizes of the freely available databases), and we restrict the
comparison to the network cost (which is hardware-independent). 

The results are reported in Table 2, as the percentage of communication that DH-MPPJ
requires compared to that of IDCloak. We observe that DH-MPPJ consistently outperforms
IDCloak, and that the gap increases with the number of parties.

### Experiments

This section covers the steps required to perform Experiments 1 and 2. We provide
estimates of the time and storage for both a `light` experiment set over a single laptop
machine and for a `full` experiment set over our experimental setup with two Docker hosts
(see **Our Setup** above). These times assume that the environment is set up and that the
images are built, according to the previously described procedure.

#### Experiment 1: End-to-End Benchmarks

- Time: 
   - `light`, single-laptop: 1 min human + 3 min compute 
   - `full`, two-server setting: 5 min human + 50 min compute
- Storage: ~600 MB of Docker images for the Docker host.

This experiment runs the DH-MPPJ protocol for an experiment grid (with several repetitions
for each value), and measures the end-to-end execution time and network cost. It is based
on a Python experiment runner, `mppj-exps/main.py`, which starts Docker containers for
each party, possibly on remote hosts. 

The script takes care of generating the test data and checks for the correctness of the
output. The script outputs some debugging information on `stderr` and the experiment
result on `stdout`, as JSON-formatted values.

Generating the results of Table 3 is a two-step process (`run_experiment_1.sh` combines
them in a single script):

1. Running the light experiment grid and storing the result in `result.json`:
```bash
python3 mppj-exps/main.py light > results.txt
```
2. Outputting a human-readable view over the `result.txt` file:
```bash
python3 mppj-exps/parse_result.py results.txt
```

**Further Configuration**

By default, the script is already pre-configured (for a localhost execution), but it can be
parameterized through various constants and runtime arguments: 

- The **experiment grids** can be specified through the `EXPERIMENT_GRIDS` constant. By
 default, it provides three grids:
    1. `light` is a reduced experiment grid for $(n,m). \in \{2,3\}\times\{1353, 1700\}$
 and only one repetition per experiment.
    2. `full` is the full experiment grid of Table 2, i.e., 
 $(n,m). \in \{2,3\}\times\{1353, 1700, 19735, 45211, 253680\}$ and does three
 repetitions per experiment.
    3. `genmpc_comp` is the experiment grid for Experiment 2.
 The script reads which grid to execute from its first positional argument.
- The **hosts** on which the various parties are run can be specified from the 
  `HELPER_HOST`, `RECEIVER_HOST`, and `SOURCES_HOSTS`. The latter is a list of hosts among
 which the sources are distributed. The default values run all parties on
  `localhost`. When running the experiment on remote Docker hosts, these values should be
 replaced by the IP addresses of the nodes.
- The **tls** configuration can be specified via the `TLS` constant. When running the
 experiments on remote Docker hosts, the config should point to the certificates
 generated during the setup (by default, in the `.cert` folder).
- The script takes an optional second integer `SKIP_TO` parameter. If provided, it skips the
 first `SKIP_TO` experiments of the experiment grid. This is meant to facilitate restarting
 an experiment that encountered an issue.


#### Experiment 2: Comparison to Generic MPC

- Time: 
   - `light`, single-laptop: 1 min human + 3 min compute 
   - `full`, server setting: 5 min human + 50 min compute
- Storage: ~4 GB of Docker images at the Docker host.

This second experiment runs both a generic MPC protocol and our DH-MPPJ implementation for
two sources and $m \in \{10, 100, 1000\}$. It supports Claim 2.

Generating the results of Table 2 is a three-step process (`run_experiment_2.sh` combines
them in a single script):

1. Run the DH-MPPJ part of the experiment:
```bash
python3 mppj-exps/main.py genmpc_comp > results-mppj.txt
```
2. Run generic MPC part (in the `light` configuration):
```bash
python3 mpspdz-exps/main.py light > results-mpc.txt
```
3. Output the results in a human-readable table:
```bash
python3 mpspdz-exps/parse_results.py results-mppj.txt results-mpc.txt
```

**Notes**
- By default, the experiment is set up to run the MP-SPDZ experiment in a single Docker
 host. This is due to the high computation and communication cost for the size 1000 join.
 Although this cannot fully represent the effect of network-introduced latency, this
 experiment is sufficient to back Claim 2.
- To run the full experiment for the MPC part (with a size 1000 dataset), use the `full`
 parameter for the `mpspdz-exps/main.py` script. 


## Limitations

**Reproducibility Limitations**

The main limiting factor for reproducing our results with the provided artifact is the
difference in the experimental setup and computing equipment. We took the following steps
to mitigate this: 

- The experiments can be run conveniently on remote servers. 
- The results in our paper are measured over standardized computing platforms (AWS EC2
 instances) that can be rented.

**General Limitation**

Our experiments also have a few limitations regarding the comparison to related works and
their implementation, which we detail below.

For the **comparison with IDCloak**, we were sadly unable to run the provided 
[source code](https://github.com/chenshuyuhhh/IDCloak.git) over a network (or even on the
`localhost` interface). Furthermore, we were unable to use the provided artifact to run a
PJC with a trivial (reconstruction-only) compute phase, in a reasonable amount of time. As
a result, we limited our comparison to the network cost reported in the IDCloak paper,
which is hardware and network-independent. Note that IDCloak is an unpublished
work at the time of this research, and this situation will probably improve in the future.

For the **comparison with generic-MPC**, we were limited by our unfamiliarity with the
MP-SPDZ framework and its compilation capabilities. We acknowledge that some fine-tuning
of the circuit and compilation option might result in a more efficient execution. However,
we do not believe that it would drastically change the comparative result. In fact,
MP-SPDZ specifically targets novices with its easy-to-use programming style and is 
optimally suited to give a rough estimate of the cost of a generic solution, which is a
sufficient level for our purposes. 

## Notes on Reusability

While the `mppj-go` code-base cannot be considered _production-ready_, our artifact
implements a complete prototype of the DH-MPPJ protocol. Notably:
- It implements the sources/helper/receiver as standalone executables which can be
 executed over a network and consume .csv input databases.
- The implementation relies only on standard Go dependencies, is cross-platform, and easy
 to build with the standard Go toolchain. 
- While we did not extract the Go package as a standalone repository, doing so would
 require minimal effort. We encourage interested users to reach out if that would be
 useful.


## Setting up remote Docker Hosts

We provide a set of utilities for setting up a set of remote Docker hosts for use in the
experiment. Although the process below should work for most setups, it might need to be
adapted to the experimenter's environment and computing infrastructure. 

This procedure and utilities assume that the experimentator machine has SSH access to all
the hosts to configure. We assume that the remote username is the same as the local one
(this is the default for Ansible but can be configured otherwise), and that the remote
user is part of `docker` group (hence can run the `docker` command without `sudo`).

From the root directory of the artifact:

1. Create an `inventory.yml` YAML file containing all nodes. The syntax is that of an
   Ansible inventory, a template is provided below, where the **IP addresses** of the node
   should be substituted:
   ```bash
      cat > setup/inventory.yml <<EOF
      all:
         hosts:
            helper:
               ansible_host: XXX.XXX.XXX.XXX
            nodes:
               ansible_host: YYY.YYY.YYY.YYY
      EOF
   ```
2. Generating the certificates for mutual authentication of the orchestrator and the
   Docker hosts:
   ```bash
      setup/gen_certs.sh setup/inventory.yml
   ```
   This generates certificates for the orchestrator (client) and each of the hosts in the
   inventory file. The script uses the hostname for the cert and generates the
   certificates as required by Docker.
3. Run the Ansible configuration script:
   ```bash
      ansible-playbook -i setup/inventory.yml setup/setup_docker_host.yml
   ```
   This installs the Docker engine on each host and configures it to accept the
   orchestrator's certificate as a client. This also clones the artifact repository.
4. Configure the network/firewall/security groups: the Docker hosts should be reachable on
   the Docker engine API port (default 2376). Moreover, the host(s) running the sources
   and receiver must reach the helper machine on the MPPJ port (default 40000). Since our
   prototype implementation does not authenticate the parties, the helper host should only
   accept connections to that port from the source/receiver hosts' IPs.
5. Finally, the required images should be built on the remote Docker hosts (see the
   Makefile targets above).

After these steps, the remote Docker hosts should be ready to run experiments. The root
artifact directory should contain a `.certs` directory with all the certificates.
