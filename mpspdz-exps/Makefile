.PHONY: mpspdz docker

MP-SPDZ:
	git submodule init && git submodule update

docker: mpspdz
	docker build --tag mpspdz:exp .

mpspdz: MP-SPDZ
	cd MP-SPDZ && \
	docker build --tag mpspdz:atlas --build-arg machine=atlas-party.x .


