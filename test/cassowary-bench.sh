#!/bin/sh
# https://github.com/rogerwelin/cassowary

set -ex

if [ -z "$TOKEN_VALUE" ]; then
	echo "Error: missing \$TOKEN_VALUE"
	exit 1
fi

CASSOWARY=${CASSOWARY:-cassowary}

exec $CASSOWARY $@ -H "Authorization: Bearer $TOKEN_VALUE"

