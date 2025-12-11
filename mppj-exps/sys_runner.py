import docker
import requests
import sys
from multiprocessing.dummy import Pool

IMAGE_NAME = "mppj"
IMAGE_TAG = "base"

CONTAINER_LABEL = "mppj-exp"
NODE_CONTAINER_PREFIX = "node"
HELPER_CONTAINER_NAME = "helper"
RECEIVER_CONTAINER_NAME = "receiver"
HELPER_PORT = 40000

FLAG_NODE_ID = "id"
FLAG_CLOUD_ADDR = "helper_address"

def host_to_docker_host_url(host):
    return 'unix://var/run/docker.sock' if host == 'localhost' else "tcp://%s:2376" % host

def log(str, end="\n"):
    print(str, file=sys.stderr, end=end, flush=True)

def dict_to_args(d):
    return " ".join(["-%s %s" % (k, v) for k, v in d.items()])

class DockerNodeSystem:
    def __init__(self, n_parties, parties_hosts, helper_host, helper_addr, receiver_host, tls, **params):
        self.N = n_parties
        self.pool = Pool(self.N)
        self.params = params

        try:
            self.docker_hosts = {host: docker.DockerClient(base_url=host_to_docker_host_url(host), tls=tls if host != 'localhost' else None) for host in set(parties_hosts+[helper_host, receiver_host])}
        except Exception as e:
            log("Error: Could create docker clients: %s\n have the certificates been created for remote hosts ?" % e)
            sys.exit(1)

        #log("Connected to docker hosts: %s" % ", ".join(self.docker_hosts.keys()))
        
        self.node_ids = ["%s%d" % (NODE_CONTAINER_PREFIX, i) for i in range(self.N)]
        self.node_host = {nid: self.docker_hosts[parties_hosts[i%len(parties_hosts)]] for i, nid in enumerate(self.node_ids)}
        
        self.helper_hostname = helper_host
        self.helper_addr =  helper_addr
        self.helper_id = HELPER_CONTAINER_NAME
        self.node_host[self.helper_id] = self.docker_hosts[helper_host]
    
        self.receiver_host = receiver_host
        self.receiver_id = RECEIVER_CONTAINER_NAME
        self.node_host[self.receiver_id] = self.docker_hosts[receiver_host] 


        self.clean_all()
        # creates a new image from the base image for each node
        def create_img(nid, host):
            try:
                host.images.remove("%s:%s" % (IMAGE_NAME, nid), force=True)
                #log("removed image %s:%s" % (IMAGE_NAME, nid))
            except requests.exceptions.HTTPError as e:
                pass  # image did not exist
            except Exception as e:
                log("could not remove image %s:%s: %s" % (IMAGE_NAME, nid, e))

            host.images.get("%s:%s" % (IMAGE_NAME, IMAGE_TAG)).tag(IMAGE_NAME, tag="%s" % nid)
            log("created %s:%s" % (IMAGE_NAME, nid))
        #self.pool.starmap(create_img, self.node_host.items())
        for nid, host in self.node_host.items():
            create_img(nid, host)
        


    def get_all(self):
        return [container 
                for host in self.docker_hosts.values() 
                for container in host.containers.list(filters={"label": CONTAINER_LABEL}, all=True, ignore_removed=True)]


    def stop_all(self):
        for container in self.get_all():
            container.kill()

    def clean_all(self):
        for container in self.get_all():
            try:
                container.remove(force=True)
                #log("removed container %s" % container.name)
            except:
                continue

    
    def run_node(self, node_id, cmd, detach=True, **args):
        args = {FLAG_NODE_ID: node_id} | args
        cmd += " " + dict_to_args(args)
        cmd = "'%s'" % cmd
        #log("running on %s: %s" % (node_id, cmd))
        container =  self.node_host[node_id].containers.run("%s:%s" % (IMAGE_NAME, node_id),
                                                name=node_id,
                                                hostname=node_id,
                                                command=cmd,
                                                network_mode="host",
                                                labels=[CONTAINER_LABEL],
                                                detach=detach)
        return container
    

    
    def run_helper(self, cmd="", detach=True, **args):
        return self.run_node(self.helper_id, cmd, detach=detach, **args)
    
    def run_receiver(self, cmd="", detach=True, **args):
        args = {FLAG_CLOUD_ADDR: self.helper_addr} | args
        return self.run_node(self.receiver_id, cmd, detach=detach, **args)

    def run_all_players(self, command,  **args):
        return self.pool.map(lambda nid: self.run_node(nid, command, **args), self.node_ids)
    
    def run_all_players_with_helper_addr(self, command, **args):
        args = {FLAG_CLOUD_ADDR: self.helper_addr} | args
        return self.pool.map(lambda nid: self.run_node(nid, command, **args), self.node_ids)

    def run_passive_players(self, command, **args):
        return self.pool.map(lambda nid: self.run_node(nid, command, **args), self.node_ids[1:])

    def run_active_player(self, command, **args):
        return self.run_node(self.node_ids[0], command, **args)

    def commit_nodes(self, node_containers):
        def commit(c):
            nid = c.name
            status = c.wait()["StatusCode"]
            if status != 0:
                raise Exception("container %s exited with status %d, logs:\n%s" % (nid, status, c.logs()))

            self.node_host[nid].images.get("%s:%s" % (IMAGE_NAME, nid)).tag("%s" % IMAGE_NAME, tag="%s-old" % nid)
            c.commit(repository="%s" % IMAGE_NAME, tag=nid)
            c.remove()
            self.node_host[nid].images.remove("%s:%s-old" % (IMAGE_NAME, nid))
            #log("committed %s" % nid)
            return 
        self.pool.map(commit, node_containers)