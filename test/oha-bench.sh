#!/bin/sh
# https://github.com/hatoo/oha

set -ex

if [ -z "$TOKEN_VALUE" ]; then
	echo "Error: missing \$TOKEN_VALUE"
	exit 1
fi

OHA=${OHA:-oha}

exec $OHA -H "Authorization: Bearer $TOKEN_VALUE" "$@"
