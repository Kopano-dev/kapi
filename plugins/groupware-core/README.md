# Kopano API groupware-core plugin

The groupware-core plugin provides the frontend HTTP proxy for the Kopano
Groupware Core REST api provided by kopano-mfr sockets.

## Kopano Groupware Core Master Fleet Runner

Kopano MFR is part of Kopano Groupware Core and needs to be started so that its
created sockets can be found and accessed by kapid.

## Configuration

`KOPANO_GC_REST_SOCKETS` is an environment variable which defines the base
directory where the Groupware Core plugin finds its required backend sockets.
All `rest*.sock` files in that directory will be used as upstream proxy paths,
and all`notify*.sock` files in that directory will be used as upstream proxy
paths for the subsription socket API. Kopano Groupware Core Master Fleet Runner
can be used to provide these sockets.

`KOPANO_GC_REST_ALLOW_CORS` is an environment variable which if set to `1`
enables CORS (Cross Origin Resource Sharing) HTTP requests and headers so that
the REST endpoints provided by this plugin can be used from a Browser cross
origin.

`KOPANO_GC_REQUIRED_SCOPES` is an environment variable which defines the
required access token scopes to grant access to the API endpoints provided by
this plugin. By default the scopes `profile, email, kopano/gc`.

## Debugging

Sometimes it is useful to see the request payload data which is sent/received
by this plugin from the upstream Kopano MFR. This can be done by using `socat`
with the sockets.

Example:
```
socat -t100 -x -v UNIX-LISTEN:/run/kopano-rest-mfr/rest0.sock,mode=777,reuseaddr,fork UNIX-CONNECT:/run/kopano-rest-mfr/rest0.sock.orig
```

The above example assumes that Kopano MFR is started and a rest socket file
has been renamed to rest0.sock.orig so `socat` can forward the traffic to it.
