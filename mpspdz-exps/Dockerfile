FROM mpspdz:atlas

COPY ./join10.mpc ./Programs/Source/join10.mpc
COPY ./join100.mpc ./Programs/Source/join100.mpc
COPY ./join1000.mpc ./Programs/Source/join1000.mpc

RUN ./compile.py join10
RUN ./compile.py join100
RUN ./compile.py join1000
