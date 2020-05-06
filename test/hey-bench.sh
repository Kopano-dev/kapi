#!/bin/sh
# https://github.com/rakyll/hey

set -ex

if [ -z "$TOKEN_VALUE" ]; then
	echo "Error: missing \$TOKEN_VALUE"
	exit 1
fi

HEY=${HEY:-hey}

exec $HEY -H "Authorization: Bearer $TOKEN_VALUE" $@
