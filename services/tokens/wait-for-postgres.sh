#!/bin/sh

set -e

# Explicitly concatenates all the array elements into a single string
cmd=$*

until psql -Atx "$APP_DSN" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing command"
exec "$cmd"
