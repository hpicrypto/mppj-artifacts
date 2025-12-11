#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 [light|full]"
    exit 1
fi

./run_experiment_1.sh $1
./run_experiment_2.sh $1

echo "Test finished! The results were saved to 'table2.txt' and 'table3.txt'."