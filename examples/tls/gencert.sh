#!/bin/bash -e

EMAIL=$(git config --get user.email 2>/dev/null || echo "someone@example.com")

cat <<_TXT_ > .ca.conf
[req]
prompt = no
distinguished_name = req_distinguished_name

[req_distinguished_name]
C = US
ST = Fake State
L = Fake Locality
O = Fake Company
# OU = Org Unit Name
# emailAddress = ${EMAIL}
CN = localhost
_TXT_

cat <<_TXT_ > .cert.conf
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
DNS.2 = 127.0.0.1
DNS.3 = ::1
# more alt. local names here
#local.dev
_TXT_

rm -f ca.pem ca.key server.key server.pem

# Generate Certificate Authority
openssl req -x509 -config .ca.conf -new -nodes -keyout ca.key -sha256 -days 1825 -out ca.pem

# Generate CA-signed Certificate
openssl genrsa -out server.key 2048
openssl req -new -config .ca.conf -key server.key -out .server.csr
openssl x509 -req -in .server.csr \
  -CA ca.pem -CAkey ca.key -CAcreateserial \
  -out server.pem -days 1825 -sha256 -extfile .cert.conf

rm .ca.conf .cert.conf .server.csr ca.srl
