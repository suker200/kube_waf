#!/bin/sh

if [ ! -f "/etc/nginx/nginx.conf" ]; then
	cp /nginx.conf /etc/nginx/nginx.conf
fi

dnsmasq

sleep 1

/usr/local/openresty/bin/openresty -c /etc/nginx/nginx.conf