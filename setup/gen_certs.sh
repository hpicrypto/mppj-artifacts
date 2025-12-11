#!/usr/bin/env bash
set -euo pipefail

# ======================================================
# Configuration
# ======================================================
INVENTORY="${1:-inventory.yml}"
DAYS=3650
DIR=".certs"


if [[ ! -f "$INVENTORY" ]]; then
    echo "Inventory file not found: $INVENTORY"
    exit 1
fi

# Convert to absolute path to avoid issues after cd
INVENTORY_ABS="$(realpath "$INVENTORY")"

# Check for yq
if ! command -v yq >/dev/null 2>&1; then
    echo "ERROR: This script requires 'yq'. Install it from https://github.com/mikefarah/yq"
    exit 1
fi

# ======================================================
# Directory setup
# ======================================================
mkdir -p "$DIR/hosts"
cd "$DIR"

echo "======================================================"
echo "🔐 Generating CA (Certificate Authority)"
echo "======================================================"

# -------------------------
# 1. Generate CA
# -------------------------
openssl genrsa -out ca-key.pem 4096

cat > ca-openssl.cnf <<EOF
[ req ]
distinguished_name = req_distinguished_name
x509_extensions = v3_ca
prompt = no

[ req_distinguished_name ]
CN = docker-ca

[ v3_ca ]
basicConstraints = critical, CA:true
keyUsage = critical, keyCertSign, cRLSign
subjectKeyIdentifier = hash
EOF

openssl req -new -x509 -days $DAYS -sha256 \
  -key ca-key.pem \
  -out ca.pem \
  -config ca-openssl.cnf

echo "CA generated: ca.pem"
echo

# ======================================================
# Extract hosts from YAML inventory
# ======================================================
echo "======================================================"
echo "📋 Parsing YAML inventory: $INVENTORY_ABS"
echo "======================================================"

HOSTS=$(yq '.all.hosts | keys | .[]' "$INVENTORY_ABS")

if [[ -z "$HOSTS" ]]; then
    echo "ERROR: No hosts found under 'all.hosts' in inventory."
    exit 1
fi

# ======================================================
# Generate server certificates
# ======================================================
for HOST in $HOSTS; do
    echo "------------------------------------------------------"
    echo "🔧 Generating certificate for host: $HOST"
    echo "------------------------------------------------------"

    IP=$(yq ".all.hosts.$HOST.ansible_host" "$INVENTORY_ABS")
    if [[ "$IP" == "null" ]]; then
        echo "ERROR: Host '$HOST' is missing ansible_host in inventory."
        exit 1
    fi

    mkdir -p "hosts/$HOST"

    echo "→ Hostname: $HOST"
    echo "→ IP:        $IP"

    # ---------------------------------------------
    # Server private key
    # ---------------------------------------------
    openssl genrsa -out "hosts/$HOST/server-key.pem" 4096

    # ---------------------------------------------
    # CSR
    # ---------------------------------------------
    openssl req -new \
      -key "hosts/$HOST/server-key.pem" \
      -subj "/CN=$HOST" \
      -out "hosts/$HOST/server.csr"

    # ---------------------------------------------
    # SAN config (hostname + IP)
    # ---------------------------------------------
    cat > "hosts/$HOST/server-openssl.cnf" <<EOF
[ v3_server ]
subjectAltName = @alt_names
extendedKeyUsage = serverAuth

[ alt_names ]
DNS.1 = $HOST
IP.1  = $IP
EOF

    # ---------------------------------------------
    # Sign server certificate
    # ---------------------------------------------
    openssl x509 -req -days $DAYS -sha256 \
      -in "hosts/$HOST/server.csr" \
      -CA ca.pem -CAkey ca-key.pem -CAcreateserial \
      -out "hosts/$HOST/server-cert.pem" \
      -extfile "hosts/$HOST/server-openssl.cnf" \
      -extensions v3_server

    echo "✔ Server certificate generated for $HOST"
    echo
done

# ======================================================
# Generate client certificate
# ======================================================
echo "======================================================"
echo "👤 Generating client certificate"
echo "======================================================"

openssl genrsa -out client-key.pem 4096

cat > client-openssl.cnf <<EOF
[ v3_client ]
extendedKeyUsage = clientAuth
EOF

openssl req -new \
  -key client-key.pem \
  -subj "/CN=docker-client" \
  -out client.csr

openssl x509 -req -days $DAYS -sha256 \
  -in client.csr \
  -CA ca.pem -CAkey ca-key.pem -CAcreateserial \
  -out client-cert.pem \
  -extfile client-openssl.cnf \
  -extensions v3_client

echo
echo "======================================================"
echo "🎉 DONE!"
echo "======================================================"
