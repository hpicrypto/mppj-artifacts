import csv
import random
import signal
import sys 
import time
import json
import gen_data
from itertools import product
from sys_runner import DockerNodeSystem
from docker.tls import TLSConfig


DEBUG = True

# ====== Experiment Grid ======
EXPERIMENT_GRIDS = {
    "light": {
        "n_sources": [2, 3, 4],
        "set_size": [1353, 1700, 19735],
        "join_size": [int(.8*1353), int(.8*1700), int(.8*19735)],
        "rep": 2
    },
    "full": {
        "n_sources": [2, 3, 4, 5, 6],
        "set_size": [1353, 1700, 19735, 45211, 253680],
        "join_size": [int(.8*1353), int(.8*1700), int(.8*19735), int(.8*45211), int(.8*253680)],
        "rep": 3
    },
    "genmpc_comp": {
        "n_sources": [2],
        "set_size": [10, 100, 1000],
        "join_size": [int(.8*10), int(.8*100), int(.8*1000)],
        "rep": 3
    }
}

# ====== Environment ======
SOURCES_HOSTS =  ["localhost"] #['node-0','node-1', 'node-2', 'node-3', 'node-4']
HELPER_HOST = "localhost"       # hostname or ip of the cloud docker host
RECEIVER_HOST = "localhost"    # hostname or ip of the receiver docker host

HELPER_ADDR = "%s:%d" % (HELPER_HOST, 40000)  # the address the sources should use to connect to the helper

# ====== Docker TLS Configuration ======
# Only used for remote docker hosts.
# Requires the certificates to be in the .certs/ folder, see setup/gen_certs.sh script,
TLS = TLSConfig(
    client_cert=('.certs/client-cert.pem', '.certs/client-key.pem'),
    ca_cert='.certs/ca.pem',
    verify=True,
) if set(SOURCES_HOSTS + [HELPER_HOST, RECEIVER_HOST]) - {'localhost'} else None

N_CPU_PER_SOURCE = 4  # 0 means all available CPUs

# Parse experiment grid from command line argument
if len(sys.argv) < 2:
    print("Usage: python main.py <experiment_grid>")
    print(f"Available grids: {list(EXPERIMENT_GRIDS.keys())}")
    sys.exit(1)

grid_name = sys.argv[1]
if grid_name not in EXPERIMENT_GRIDS:
    print(f"Error: Unknown experiment grid '{grid_name}'")
    print(f"Available grids: {list(EXPERIMENT_GRIDS.keys())}")
    sys.exit(1)


# Set experiment parameters from selected grid
grid = EXPERIMENT_GRIDS[grid_name]
N_PARTIES = grid["n_sources"]
SET_SIZE = grid["set_size"]
JOIN_SIZES = grid["join_size"]
N_REP = grid["rep"]

# Parse optional SKIP_TO parameter from command line
SKIP_TO = 0
if len(sys.argv) >= 3:
    try:
        SKIP_TO = int(sys.argv[2])
    except ValueError:
        print(f"Error: SKIP_TO parameter must be an integer, got '{sys.argv[2]}'")
        sys.exit(1)


def log(str, end="\n"):
    if DEBUG:
        print(str, file=sys.stderr, end=end, flush=True)



def get_expected_result(source_ids, m, joinsize, seed):
    random.seed(seed)
    tables = gen_data.generate_tables(source_ids, m)
    tables = gen_data.postprocess_tables(tables, joinsize)
    return {','.join(val) for val in gen_data.join_tables(tables).values()} # join all columns for set comparison

def read_node_output(node):
    reader = csv.reader((l.decode('utf-8') for l in node.logs(stdout=True, stderr=False, stream=True)))
    next(reader) # skip header
    result = set()
    for row in reader:
        result.add(','.join(row)) # join all columns for set comparison
    return result

def read_node_stats(node):
    for li in node.logs(stdout=False, stderr=True).decode('utf-8').splitlines():
        if "Stats:" in li: 
            return json.loads(li.split("Stats:")[1].strip())

def collect_stats(nodes):
    return {node.name: read_node_stats(node) for node in nodes}

def sig_handler(sig, frame):
    if sig == signal.SIGINT:
        log("Caught SIGINT, cleaning and exiting...")
        system.clean_all()
        sys.exit(0)
    if sig == signal.SIGTSTP:
        log("Caught SIGTSTP, exiting. without cleaning...")
        sys.exit(0)


# ====== Main ======
signal.signal(signal.SIGINT, sig_handler)
signal.signal(signal.SIGTSTP, sig_handler)


exps_to_run = list(product(N_PARTIES, zip(SET_SIZE, JOIN_SIZES), range(N_REP)))
log("%d experiments to run" % len(exps_to_run))

#filepath = setup_experiment_file()

for (i, exp) in enumerate(exps_to_run):

    if i+1 < SKIP_TO:
        continue

    n_party, (set_size, join_size), rep = exp

    log("======= starting experiment N=%d SetLogSize=%s JoinSize=%s REP=%d =======" % (n_party, set_size, join_size, rep))

    system = DockerNodeSystem(n_party, SOURCES_HOSTS, HELPER_HOST, HELPER_ADDR, RECEIVER_HOST, tls=TLS, set_size=set_size)

    time.sleep(1) # lets the thing clean

    source_ids = system.node_ids
    sources_id_list = ','.join(source_ids)
    seed = i

    log("generating test data...")
    all = system.run_all_players(f"gen_data.py -source_ids {sources_id_list} -generate {set_size} -joinsize {join_size} -seed {seed} >> data.csv")
    system.commit_nodes(all)
    log("data generated")
    
    log("starting helper...")
    cloud = system.run_helper(cmd="helper", sources=sources_id_list, n_rows=set_size)
    time.sleep(3)   # wait for helper to start
    log("running receiver and sources")
    timestart = time.time()

    rec = system.run_receiver("receiver", sources=sources_id_list)
    all = system.run_all_players_with_helper_addr("cat data.csv | source", n_cpu=N_CPU_PER_SOURCE)
    for l in cloud.logs(stderr=True, stdout=True, stream=True):
        log("helper%s" % l.decode('utf-8').strip("\n"))
    
    cloud.wait()
    if not rec.wait()["StatusCode"] == 0:
        log("ERROR: node returned a non-zero exit code")
        sys.exit(1)

    timeend = time.time() # gives a rough estimate but is imprecise (waits for helper to shut down, docker etc)

    log("done in %f seconds, checking results..." % (timeend - timestart))
    result_set = read_node_output(rec)
    expected_set = get_expected_result(source_ids, set_size, join_size, seed)
    if result_set != expected_set:
        log("ERROR: result does not match expected result")
        print("Expected:", expected_set)
        print("Got:", result_set)
        sys.exit(1)
    else:
        log("result OK")

    nodes_stats = collect_stats([all[0], cloud, rec])
    log({"n_sources": n_party, "set_size": set_size, "join_size": join_size, "nodes_stats": nodes_stats})


    result = {
    "n_sources": n_party,
    "set_size": set_size,
    "join_size": join_size,
    "rep": rep,
    "nodes_stats": nodes_stats,
    }

    print(json.dumps(result), flush=True)
    log("")

time.sleep(2) # some time to clean