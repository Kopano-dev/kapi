# Kopano Groupware API (grapi) plugin

The grapi plugin provides the frontend HTTP proxy for the Kopano Groupware REST
API provided by grapi (via kopano-mfr.py) sockets.

## Kopano Groupware Master Fleet Runner (grapi)

Grapi is part of Kopano Groupware and needs to be started so that its created
sockets can be found and accessed by kapid.

## Configuration

`KOPANO_GRAPI_SOCKETS` is an environment variable which defines the base
directory where the Grapi plugin finds its required backend sockets. All
`rest*.sock` files in that directory will be used as upstream proxy paths,
and all `notify*.sock` files in that directory will be used as upstream proxy
paths for the subsription socket API. Kopano Groupware Master Fleet Runner
(grapi) can be used to provide these sockets.

`KOPANO_GRAPI_ALLOW_CORS` is an environment variable which if set to `1`
enables CORS (Cross Origin Resource Sharing) HTTP requests and headers so that
the REST endpoints provided by this plugin can be used from a Browser cross
origin.

`KOPANO_GRAPI_REQUIRED_SCOPES` is an environment variable which defines the
required access token scopes to grant access to the API endpoints provided by
this plugin. By default the scopes are `profile, email, kopano/gc`.

## HTTP API v1

The base URL to this API is `/api/gc/v1`. All example URLs are sub paths of
this base URL.

### Authentication

All endpoints of the grapi API require [OAuth2 Bearer authentication](https://tools.ietf.org/html/rfc6750#section-2.1) with an access
token (if not stated otherwise).

## Debugging

Sometimes it is useful to see the request payload data which is sent/received
by this plugin from the upstream Kopano MFR. This can be done by using `socat`
with the sockets.

Example:
```
socat -t100 -x -v UNIX-LISTEN:/run/kopano-grapi/rest0.sock,mode=777,reuseaddr,fork UNIX-CONNECT:/run/kopano-grapi/rest0.sock.orig
```

The above example assumes that Kopano Grapi is started and a rest socket file
has been renamed to rest0.sock.orig so `socat` can forward the traffic to it.
