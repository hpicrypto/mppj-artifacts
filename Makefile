.PHONY: mppj-docker mppj-exps mpspdz-exps all

mppj-docker:
	cd mppj-go && make docker

mppj-exps: mppj-docker
	cd mppj-exps && make docker

mpspdz-exps:
	cd mpspdz-exps && make docker

all: mppj-exps mpspdz-exps