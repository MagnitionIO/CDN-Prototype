#!/bin/sh

# Host and port where Varnish caches are running
VARNISH_HOST1="frontend1"
VARNISH_PORT1="80"

VARNISH_HOST2="frontend2"
VARNISH_PORT2="80"

ATS_HOST1="edge-backend"
ATS_PORT1="8080"

# Loop until Varnish1 is reachable
until nc -z $VARNISH_HOST1 $VARNISH_PORT1; do
  echo "$(date) - waiting for Varnish service at $VARNISH_HOST1:$VARNISH_PORT1..."
  sleep 5
done

echo "$(date) - Varnish service is ready at $VARNISH_HOST1:$VARNISH_PORT1"

# Loop until Varnish2 is reachable
until nc -z $VARNISH_HOST2 $VARNISH_PORT2; do
  echo "$(date) - waiting for Varnish service at $VARNISH_HOST2:$VARNISH_PORT2..."
  sleep 5
done

echo "$(date) - Varnish service is ready at $VARNISH_HOST2:$VARNISH_PORT2"

# Loop until ATS is reachable
until nc -z $ATS_HOST1 $ATS_PORT1; do
  echo "$(date) - waiting for ATS service at $ATS_HOST1:$ATS_PORT1..."
  sleep 5
done

echo "$(date) - ATS service is ready at $ATS_HOST1:$ATS_PORT1"

# Start the main application
exec "$@"
