import sys
import json
import pandas as pd

# load the json file from the file path given as command line argument
if len(sys.argv) < 2:
    print("Usage: python parse_result.py <experiment_file>")
    sys.exit(1) 

experiment_file = sys.argv[1]
data = []
with open(experiment_file, "r") as f:
    for line in f:
        data.append(json.loads(line))

df = pd.json_normalize(data, sep='_')

# averages over the reps 
grouped = df.groupby(["n_sources", 'set_size', 'join_size']).mean().reset_index()

# extract the total time as the reciever time_total, in seconds
grouped['total_time'] = grouped['nodes_stats_receiver_time_total'] / 1e9  # convert to seconds

# extract the total communication as the sum of data sent and received by the helper node in MB
grouped['total_communication'] = (grouped['nodes_stats_helper_data_sent'] + grouped['nodes_stats_helper_data_recv']) / (1024 * 1024)

# drops the other columns
grouped = grouped[["n_sources", 'set_size', 'total_time', 'total_communication']]

# pivot the set size as a column
grouped = grouped.pivot(index="n_sources", columns='set_size', values=['total_time', 'total_communication'])

# print the result
final_result = grouped.to_string(
    max_colwidth=None,
    max_rows=None,
    line_width=None
) # prevents truncation in the output files
print(final_result)
