#!/usr/bin/env python3
# A script to either split a CSV file into multiple dictionaries or generate synthetic data.
# It can also ensure that only a specified number of keys are common across all dictionaries.
# Finally, it can output a specific dictionary or the join of all dictionaries on their common keys

import csv
import sys
import argparse
import random


COL_KEY = "uid"
COL_VALUE = "val"

def tables_from_csv(n_parties, input_file): # TODO adapt to source ids
    """Read a CSV file and split it into n_parties dictionaries based on columns."""
    with open(input_file, newline='') as csvfile:
        reader = csv.reader(csvfile)
        header = next(reader)
        n_attr = len(header) - 1  # excluding the key column
        if n_attr < n_parties:
            raise ValueError(f"CSV has only {n_attr} attribute columns, but {n_parties} parties were specified.")
        
        dicts = [{} for _ in range(n_parties)]

        for row in reader:
            k, attrs = row[0], row[1:]
            for i in range(n_parties): # TODO: only uses first n_parties columns
                dicts[i][k] = attrs[i]
    return dicts

def generate_tables(source_ids, n_rows):
    """Generate n_tables dictionaries each with n_rows entries."""
    n_tables = len(source_ids)
    dicts = [{} for _ in range(n_tables)]
    
    for i in range(n_rows):
        key = f"key_{i}"
        for j in range(n_tables):
            # Generate random values for each table
            value = f"val_{source_ids[j]}_{random.randint(1, 1000)}"
            dicts[j][key] = value
    
    return dicts

def postprocess_tables(dicts, joinsize):
    """Postprocess dictionaries to ensure only joinsize keys are common across all dicts."""
    if not dicts or joinsize <= 0:
        return dicts
    
    # Find all keys that appear in all dictionaries
    common_keys = sorted(list(dicts[0].keys()))

    if len(common_keys) <= joinsize:
        raise ValueError("joinsize is larger than the number of common keys available.")

    # Select joinsize random keys to keep common
    keys_to_remove = random.sample(common_keys, len(common_keys) - joinsize)

    # For each key to remove from the join, randomly choose which dictionaries to remove it from
    # (keep it in at least one dict, remove from others), and change it to a non-conflicting name.
    for key in keys_to_remove:
        dicts_to_modify = random.sample(dicts, random.randint(1, len(dicts)-1))
        newkey = f"removed_{key}"
        for dict in dicts_to_modify:
            dict[newkey] = dict[key]
            del dict[key]
            
    return dicts


def join_tables(dicts):
    """Join all tables on common keys."""
    if not dicts:
        return {}
    
    # Find keys that exist in all dictionaries
    common_keys = set(dicts[0].keys())
    for d in dicts[1:]:
        common_keys &= set(d.keys())
    
    # Build joined table
    joined = {}
    for key in common_keys:
        values = [d[key] for d in dicts]
        joined[key] = values
    
    return joined

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Split CSV into multiple dictionaries or generate synthetic data')
    
    # Create mutually exclusive group for input source
    input_group = parser.add_mutually_exclusive_group(required=True)
    input_group.add_argument('-file', type=str, help='Input CSV file to process')
    input_group.add_argument('-generate', type=int, help='the number of rows to generate')
    
    parser.add_argument('-joinsize', type=int, help='Number of keys that should be common across all dictionaries')
    parser.add_argument('-seed', type=int, help='Random seed for reproducible results')
    parser.add_argument('-source_ids', type=str, required=True, help='Comma-separated list of source IDs for tables')
    parser.add_argument('-id', type=str, required=True, help='The id of the source to output, any value that is not a source id will output the join of all tables')
    
    args = parser.parse_args()
    
    # Create mapping from source IDs to table indices
    source_ids = [sid.strip() for sid in args.source_ids.split(',')]
    n_parties = len(source_ids)
    source_id_to_index = {sid: idx for idx, sid in enumerate(source_ids)}
    
    if args.seed is not None:
        random.seed(args.seed)
    
    if args.file:
        result_dicts = tables_from_csv(n_parties, args.file)
    else:  # args.generate
        n_rows = args.generate
        result_dicts = generate_tables(source_ids, n_rows)
    
    if args.joinsize is not None:
        result_dicts = postprocess_tables(result_dicts, args.joinsize)
    

    # Handle output selection
    if args.id in source_id_to_index:
        table_idx = source_id_to_index[args.id]
        d = result_dicts[table_idx]
        print(f"{COL_KEY},{COL_VALUE}")
        for key in sorted(d.keys()):
            print(f"{key},{d[key]}")
    else:
        joined_result = join_tables(result_dicts)
        print(','.join([f"{COL_VALUE}-{sid}" for sid in source_ids]))
        for key in sorted(joined_result.keys()):
            print(f"{key},{','.join(joined_result[key])}")

