#!/bin/bash

CERTNAME=$1
DOMAIN=$2

openssl req -nodes -x509 -newkey rsa:4096 -keyout ${CERTNAME}_KEY.pem -out ${CERTNAME}.pem -days 365 -subj "/CN=${DOMAIN}i,/OU=${DOMAIN}-OU"


## sh scripts/gen_cert.sh tester_cert tester.suker200.com

cat ${CERTNAME}_KEY.pem > cert.pem
cat ${CERTNAME}.pem >> cert.pem

# kubectl create secret generic db-user-pass --from-file=cert=cert.pem 

unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     os=Linux;;
    Darwin*)    os=Mac;;
    *)          os="UNKNOWN:${unameOut}"
esac

if [ "$os" == "Mac" ]; then
	sed -i ''  "s/REPLACE_CERT_CONTENT/$(cat cert.pem | base64)/" cert_secret.yaml
else
	sed "s/REPLACE_CERT_CONTENT/$(cat cert.pem | base64)/" cert_secret.yaml -i
fi

