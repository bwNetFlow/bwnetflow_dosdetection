#!/bin/bash

openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:4096 -out ./ssl/private.key
openssl req -new -x509 -sha256 -key ./ssl/private.key -out ./ssl/server.crt -days 365
