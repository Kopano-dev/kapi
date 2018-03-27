# Source me :)

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

response=$(curl -sSk $tokenEndpoint --data "grant_type=refresh_token&refresh_token=$REFRESH_TOKEN_VALUE&client_id=$CLIENT_ID")
echo $response | jq .
accessTokenValue=$(echo "$response"|jq -r .access_token)
if [ -z "$accessTokenValue" -o "$accessTokenValue" = "null" ]; then
	echo "Error: failed to retrieve access token - \$REFRESH_TOKEN_VALUE invalid?"
	return
fi

echo TOKEN_VALUE=$accessTokenValue
export TOKEN_VALUE=$accessTokenValue
