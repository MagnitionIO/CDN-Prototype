#!/bin/sh

# L1 CACHES
L1_1_HOST="frontend1"
L1_1_PORT="80"

L1_2_HOST="frontend2"
L1_2_PORT="80"

L1_3_HOST="frontend3"
L1_3_PORT="80"

L1_4_HOST="frontend4"
L1_4_PORT="80"

# L2 CACHES
L2_1_HOST="edge-backend-1"
L2_1_PORT="80"

L2_2_HOST="edge-backend-2"
L2_2_PORT="80"

L2_3_HOST="edge-backend-3"
L2_3_PORT="80"

L2_4_HOST="edge-backend-4"
L2_4_PORT="80"

# L1 CACHES
until nc -z $L1_1_HOST $L1_1_PORT; do
  echo "$(date) - waiting for service at $L1_1_HOST:$L1_1_PORT..."
  sleep 2
done

echo "$(date) - service is ready at $L1_1_HOST:$L1_1_PORT"

until nc -z $L1_2_HOST $L1_2_PORT; do
  echo "$(date) - waiting for service at $L1_2_HOST:$L1_2_PORT..."
  sleep 2
done

echo "$(date) - service is ready at $L1_2_HOST:$L1_2_PORT"

until nc -z $L1_3_HOST $L1_3_PORT; do
  echo "$(date) - waiting for service at $L1_3_HOST:$L1_3_PORT..."
  sleep 2
done

echo "$(date) - service is ready at $L1_3_HOST:$L1_3_PORT"

until nc -z $L1_4_HOST $L1_4_PORT; do
  echo "$(date) - waiting for service at $L1_4_HOST:$L1_4_PORT..."
  sleep 2
done

echo "$(date) - service is ready at $L1_4_HOST:$L1_4_PORT"


# L2 CACHES
until nc -z $L2_1_HOST $L2_1_PORT; do
  echo "$(date) - waiting for service at $L2_1_HOST:$L2_1_PORT..."
  sleep 2
done

echo "$(date) - service is ready at $L2_1_HOST:$L2_1_PORT"

until nc -z $L2_2_HOST $L2_2_PORT; do
  echo "$(date) - waiting for service at $L2_2_HOST:$L2_2_PORT..."
  sleep 2
done

echo "$(date) - service is ready at $L2_2_HOST:$L2_2_PORT"

until nc -z $L2_3_HOST $L2_3_PORT; do
  echo "$(date) - waiting for service at $L2_3_HOST:$L2_3_PORT..."
  sleep 2
done

echo "$(date) - service is ready at $L2_3_HOST:$L2_3_PORT"

until nc -z $L2_4_HOST $L2_4_PORT; do
  echo "$(date) - waiting for service at $L2_4_HOST:$L2_4_PORT..."
  sleep 2
done

echo "$(date) - service is ready at $L2_4_HOST:$L2_4_PORT"

# Start the main application
exec "$@"
