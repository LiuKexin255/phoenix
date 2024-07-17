#!/bin/bash
set -ex

# Generate a server cert.
openssl genrsa -out server.key 4096  

openssl req -utf8 -new \
    -config openssl.conf \
    -out server.req \
    -key server.key \
    -subj /C=CN/ST=Guangdong/L=Shenzhen/O=gRPC/CN=*.liukexin.com/

openssl x509 -req -sha384 \
    -extfile openssl.conf \
    -extensions v3_req \
    -in server.req \
    -out server.cer \
    -CAkey https_ca.key \
    -CA https_ca.cer  \
    -days 36500 \
    -CAcreateserial -CAserial serial

rm *.req