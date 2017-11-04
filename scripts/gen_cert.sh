#!/bin/bash

CERTNAME=$1
DOMAIN=$2

openssl req -nodes -x509 -newkey rsa:4096 -keyout ${CERTNAME}_KEY.pem -out ${CERTNAME}.pem -days 365 -subj "/CN=${DOMAIN}i,/OU=${DOMAIN}-OU"


## sh scripts/gen_cert.sh tester_cert tester.suker200.com