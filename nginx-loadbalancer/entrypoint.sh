#!/bin/sh

nohup /postStart.sh &

/usr/local/openresty/bin/openresty -g 'daemon off;'