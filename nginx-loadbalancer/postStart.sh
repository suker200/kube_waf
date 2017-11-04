#!/bin/sh

while true
do
	curl -s localhost:10254/healthz | grep "ok"
	if [ $? -eq 0 ]; then # Backend okie --> Ready 
		if [ -f "/tmp/terminated" ]; then # Terminated pod , trigger by preStop from k8S
			rm -f /var/www/html/check
			/usr/local/openresty/bin/openresty -s reload
			exit
		fi

		if [ ! -d "/var/www/html" ]; then
			mkdir -p /var/www/html
		fi

		if [ ! -f /var/www/html/check ]; then
			echo 'hello' > /var/www/html/check
			/usr/local/openresty/bin/openresty -s reload
		fi
		echo 'hello' > /var/www/html/check
	else # Backend not okie --> notReady
		if [ -d "/var/www/html" ]; then
			rm -f /var/www/html/check
		fi
	fi
	sleep 30
done