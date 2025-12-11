#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 [light|full]"
    exit 1
fi

python3 mppj-exps/main.py genmpc_comp > results-mppj.txt && \
python3 mpspdz-exps/main.py $1 > results-mpc.txt && \
python3 mpspdz-exps/parse_results.py results-mppj.txt results-mpc.txt | tee table2.txt && \
echo "result table saved as 'table2.txt'."