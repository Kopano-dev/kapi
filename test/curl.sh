#!/bin/sh

# Simple curl wrapper which injects TOKEN_VALUE from environment into the
# Authentication request header as Bearer auth.

set -ex

if [ -z "$TOKEN_VALUE" ]; then
	echo "No \$TOKEN_VALUE - check your env"
	exit 1
fi

exec curl -H "Authorization: Bearer $TOKEN_VALUE" "$@"
