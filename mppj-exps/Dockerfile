FROM mppj:latest

# Python3 is needed to generate the test data
RUN apt-get update
RUN apt-get install -y python3

# Copy the data generation script
COPY --chmod=700 ./gen_data.py /usr/local/bin/

# Set entrypoint to allow specifying which command to run
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["echo 'Please specify a command to run'"]