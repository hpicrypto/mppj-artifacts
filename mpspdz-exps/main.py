import json
import docker
import sys

from docker.tls import TLSConfig

def log(str, end="\n"):
    print(str, file=sys.stderr, end=end, flush=True)

DOCKER_HOST = 'localhost'

DOCKER_IMAGE = "mpspdz:exp"

# the first argument should be either "light" or "full" for the experiment grid
if len(sys.argv) < 2 or sys.argv[1] not in ["light", "full"]:
    log("Usage: python main.py [light|full]")
    sys.exit(1)

ROWS = [10, 100] if sys.argv[1] == "light" else [10, 100, 1000]

TLS = None
if DOCKER_HOST != 'localhost':
    TLS = TLSConfig(
        client_cert=('.certs/client-cert.pem', '.certs/client-key.pem'),
        ca_cert='.certs/ca.pem',
        verify=True,
    )

docker_host = docker.DockerClient(base_url='unix://var/run/docker.sock' if DOCKER_HOST == 'localhost' else "tcp://%s:2376" % DOCKER_HOST, tls=TLS)


for rows in ROWS:
    program = f"join{rows}"
    log(f"Running program: {program}")

    container = docker_host.containers.run(
        image=DOCKER_IMAGE,
        command=f"./Scripts/compile-run.py atlas {program}",
        detach=True,    
    )

    for line in container.logs(stream=True):
        ll = line.strip().decode('utf-8')
        log(ll)
        # Parses `Time = X.XXXX seconds`
        if ll.startswith("Time ="):
            time = float(ll.split('=')[1].strip().split(' ')[0])
        # Parses `Global data sent = X.XXXX MB`
        if ll.startswith("Global data sent ="):
            comm = float(ll.split('=')[1].strip().split(' ')[0])

    if not time or not comm:
        log("Error: could not parse time or communication from container logs")
        sys.exit(1)

    print(json.dumps({
        "n_sources": 2,
        "set_size": rows,
        "total_time": time,
        "total_communication": comm,
    }), flush=True)