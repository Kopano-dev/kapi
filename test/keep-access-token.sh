# Source me :)

# Simple script to enrich the shell environment with a Kopano Konnect access
# token and id tokens from an refresh token which already is in the environment.
#
# It works best with a local .auth file which is an file containing a bunch of
# environment variables and is sourced before running this script. It looks like
# this:
#
#  ````
#  export ISS=https://your-kopano.local
#  export CLIENT_ID=my-client-id
#  export REFRESH_TOKEN_VALUE=replace-with-refresh-token-value
#  ````
#
# If you do not have a local .auth file or a REFRESH_TOKEN yet, check the
# `get-access-token.py` script. Remember to request the `offline_access` scope
# to get an refresh token.
#
# Runtime dependencies:
#  - curl
#  - jq
#
# Environment variables supported:
#
#  ISS            : OpenID Connect Identifier
#  CLIENT_ID      : Client ID as known to the OpenID Connect Identifier
#  CLIENT_SECRET  : Client secret for client ID (optional)
#  REFRESH_TOKEN  : Refresh Token from the OpenID Connect Identifier
#
# Note that the REFRESH_TOKEN must be issues must have a matching for the
# CLIENT_ID value.
#
# You can use the `curl.sh` or `oidc-pyko-console.py` scripts to make use of the
# TOKEN_VALUE exported by this script and to access Kopano services protected by
# access token.

if [ -z "$ISS" ]; then
    echo "No \$ISS - check your env"
    return
fi

if [ -z "$REFRESH_TOKEN_VALUE" ]; then
    echo "No \$REFRESH_TOKEN_VALUE - check your env"
    return
fi

if [ -z "$CLIENT_ID" ]; then
    CLIENT_ID=playground.js
fi

tokenEndpoint=$(curl -sSk $ISS/.well-known/openid-configuration|jq -r .token_endpoint)
if [ -z "$tokenEndpoint" ]; then
	echo "Error: failed to discover token endpoint"
	return
fi

response=$(curl -sSk $tokenEndpoint --data "grant_type=refresh_token&refresh_token=$REFRESH_TOKEN_VALUE&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET")
echo $response | jq .
accessTokenValue=$(echo "$response"|jq -r .access_token)
if [ -z "$accessTokenValue" -o "$accessTokenValue" = "null" ]; then
	echo "Error: failed to retrieve access token - \$REFRESH_TOKEN_VALUE invalid?"
	return
fi

echo TOKEN_VALUE=$accessTokenValue
export TOKEN_VALUE=$accessTokenValue
