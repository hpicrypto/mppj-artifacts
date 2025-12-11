FROM golang:1.24 AS builder

WORKDIR /app

# Copy go mod files
COPY ./go.mod ./go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all commands in cmd folder
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -v -o . ./cmd/...

FROM ubuntu:latest

# Copy built binaries from builder stage
COPY --from=builder /app/source /usr/local/bin/
COPY --from=builder /app/helper /usr/local/bin/
COPY --from=builder /app/receiver /usr/local/bin/

# Set entrypoint to allow specifying which command to run
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["echo 'Please specify a command to run'"]