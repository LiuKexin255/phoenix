#!/bin/bash
set -ex

openssl genrsa -out ca.key 4096

# Create the server CA certs.
openssl req -x509 -utf8 -new                 \
  -days 3650                                 \
  -key ca.key                                \
  -out ca.cer                                \
  -subj /C=CN/ST=Guangdong/L=Shenzhen/O=gRPC/CN=liukexin-grpc/