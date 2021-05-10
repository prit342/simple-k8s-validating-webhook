#!/usr/bin/env bash

# Generate the CA cert and private key for our self signed cert
openssl req -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=Webhook Demo CA"
## Generate the private key for the webhook server

openssl genrsa -out webhook-server-tls.key 2048

openssl req -newkey rsa:2048 -nodes -keyout webhook-server-tls.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=webhook-server.webhook-demo.svc" -out webhook-server-tls.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:webhook-server.webhook-demo.svc,DNS:webhook-server.webhook-demo,DNS:webhook-server") -days 365 -in webhook-server-tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out webhook-server-tls.crt
