.PHONY: MP-SPDZ mpspdz docker

MP-SPDZ:
	git submodule update --init

docker: mpspdz
	docker build --tag mpspdz:exp .

mpspdz: MP-SPDZ
	cd MP-SPDZ && \
	docker build --tag mpspdz:atlas --build-arg machine=atlas-party.x .


