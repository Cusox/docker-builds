#!/bin/bash

set -e

CERT_DIR="./certs"
DERP_HOST=${1:="derp.local"}

echo "Generating self-signed certificate..."

if [ ! -d "$CERT_DIR" ]; then
    mkdir -p "$CERT_DIR"
fi

echo "Generating private key..."

openssl genrsa -out "$CERT_DIR/$DERP_HOST.key" 2048

openssl req -new \
	-key "$CERT_DIR/$DERP_HOST.key" \
	-out "$CERT_DIR/$DERP_HOST.csr" \
	-subj "/CN=$DERP_HOST" \
	-addext "subjectAltName=DNS:${DERP_HOST}"

openssl x509 -req \
	-days 36500 \
	-in "$CERT_DIR/$DERP_HOST.csr" \
	-signkey "$CERT_DIR/$DERP_HOST.key" \
	-out "$CERT_DIR/$DERP_HOST.crt" \
	-extfile <(printf "subjectAltName=DNS:${DERP_HOST}")

echo "Certificate and key generated at $CERT_DIR/$DERP_HOST.crt and $CERT_DIR/$DERP_HOST.key"

echo "Done."
