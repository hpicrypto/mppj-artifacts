#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 [light|full]"
    exit 1
fi

python3 mppj-exps/main.py $1 > results.txt && \
python3 mppj-exps/parse_result.py results.txt | tee table3.txt && \
echo "result table saved as 'table3.txt'."