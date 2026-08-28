#!/bin/bash
set -euo pipefail

# ═══════════════════════════════════════════════════════════════════════════════
# Generate mTLS certificates for CP'S Enterprise DCS
# ═══════════════════════════════════════════════════════════════════════════════

CERT_DIR="$(dirname "$0")/certs"
CA_DIR="$(dirname "$0")/ca"

mkdir -p "$CERT_DIR" "$CA_DIR"

echo "Generating CA certificate..."
openssl req -x509 -newkey rsa:4096 -keyout "$CA_DIR/ca.key" -out "$CA_DIR/ca.crt" \
  -days 3650 -nodes -subj "/CN=dcs-internal-ca/O=CP'S Enterprise/C=SA"

echo "Generating service certificates..."
for service in regional-agent master-agent local-agent; do
  echo "  -> $service"
  openssl req -newkey rsa:2048 -keyout "$CERT_DIR/$service.key" -out "$CERT_DIR/$service.csr" \
    -subj "/CN=$service/O=CP'S Enterprise/C=SA" -nodes
  openssl x509 -req -in "$CERT_DIR/$service.csr" -CA "$CA_DIR/ca.crt" -CAkey "$CA_DIR/ca.key" \
    -CAcreateserial -out "$CERT_DIR/$service.crt" -days 365
done

echo "Certificate generation complete."
echo "CA: $CA_DIR/ca.crt"
echo "Certificates: $CERT_DIR/"
