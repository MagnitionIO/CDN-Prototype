#!/bin/bash
# /usr/local/bin/wait-for-it.sh", "edge-backend:8080 --

varnishd -F -f /etc/varnish/default.vcl -s malloc,256m -a :10080 &
varnishncsa -F '%h %l %u %t "%r" %s %b "%{Referer}i" "%{User-agent}i" "%{X-Cache}o"' -w /Users/yazhuo/Workspace/CDN-prototype-priv/access.log &
wait