.PHONY: mppj-exps mpspdz-exps all

mppj-exps:
	cd mppj-exps && make docker

mpspdz-exps:
	cd mpspdz-exps && make docker

all: mppj-exps mpspdz-exps