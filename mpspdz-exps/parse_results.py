import sys
import json
import pandas as pd

# load the json file from the file path given as command line argument
if len(sys.argv) < 3:
    print("Usage: python parse_result.py <MPPJ experiment_file> <MPC experiment_file>")
    sys.exit(1) 

mppj_experiment_file = sys.argv[1]
mpc_experiment_file = sys.argv[2]
mppj_data = []
with open(mppj_experiment_file, "r") as f:
    for line in f:
        mppj_data.append(json.loads(line))

df = pd.json_normalize(mppj_data, sep='_')

# averages over the reps 
grouped = df.groupby(["n_sources", 'set_size']).mean().reset_index()

# extract the total time as the reciever time_total, in seconds
grouped['total_time'] = grouped['nodes_stats_receiver_time_total'] / 1e9  # convert to seconds

# extract the total communication as the sum of data sent and received by the helper node in MB
grouped['total_communication'] = (grouped['nodes_stats_helper_data_sent'] + grouped['nodes_stats_helper_data_recv']) / (1024 * 1024)

# drops the other columns
grouped = grouped[['set_size', 'total_time', 'total_communication']]

# uses set_size as index
grouped = grouped.set_index('set_size')

# print the result
#print(grouped)

mpc_data = []
with open(mpc_experiment_file, "r") as f:
    for line in f:
        mpc_data.append(json.loads(line))

df_mpc = pd.json_normalize(mpc_data, sep='_')

df_mpc = df_mpc[['set_size', 'total_time', 'total_communication']]

df_mpc = df_mpc.set_index('set_size')


# merge the two dataframes on set_size
merged = pd.merge(df_mpc, grouped, left_index=True, right_index=True, suffixes=('_mpc', '_mppj'))

final_result = merged.to_string(
    max_colwidth=None,
    max_rows=None,
    line_width=None
) # prevents truncation in the output files
print(final_result)
