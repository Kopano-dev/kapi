#!/bin/sh

set -ex

if [ -z "$TOKEN_VALUE" ]; then
	echo "Error: missing \$TOKEN_VALUE"
	exit 1
fi

exec hey -H "Authorization: Bearer $TOKEN_VALUE" $@
