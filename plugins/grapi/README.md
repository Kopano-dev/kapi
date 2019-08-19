# Kopano Groupware API (GRAPI) plugin

The grapi plugin provides the frontend HTTP proxy for the Kopano Groupware REST
API provided by GRAPI (via mfr.py) sockets.

## Kopano Groupware Master Fleet Runner (GRAPI)

GRAPI is required and needs to be started, so that its created sockets can be
found and accessed by kapid. See the [GRAPI](https://stash.kopano.io/projects/KC/repos/grapi) for
details.

## Configuration

`KOPANO_GRAPI_SOCKETS` is an environment variable which defines the base
directory where the Grapi plugin finds its required backend sockets. All
`rest*.sock` files in that directory will be used as upstream proxy paths,
and all `notify*.sock` files in that directory will be used as upstream proxy
paths for the subsription socket API. Kopano Groupware Master Fleet Runner
(GRAPI) can be used to provide these sockets.

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

All endpoints of the GRAPI API require [OAuth2 Bearer authentication](https://tools.ietf.org/html/rfc6750#section-2.1) with an access
token which includes (if not specified otherwise) at least all of the scopes as
defined in the `KOPANO_GRAPI_REQUIRED_SCOPES` environment setting (see above).

### Some quick start commandline examples

Here are some examples to give you the idea how the access the GRAPI REST
endpoints. The examples use the `./test/curl.sh` script which automatically
injects a valid access token (from environment) into the request.

```
./test/curl.sh https://kopano-dev.local/api/gc/v1/me
{
  "mobilePhone":"(039) 6781727",
  "@odata.context":"\/api\/gc\/v1\/me",
  "surname":"Peters",
  "displayName":"Marijn Peters",
  "id":"AAAAAKwhqVBA0-5Isxn7p1MwRCUBAAAABgAAAG0zAABNdz09AAAAAA==",
  "jobTitle":"Operations geologist",
  "userPrincipalName":"user3",
  "officeLocation":"Delft",
  "mail":"user3@kopano-dev.local",
  "givenName":"Marijn"
}%
```

```
./test/curl.sh https://kopano-dev.local/api/gc/v1/users/user1
{
  "mobilePhone":"05416 169130",
  "@odata.context":"\/api\/gc\/v1\/users\/user1",
  "surname":"Brekke",
  "displayName":"Jonas Brekke",
  "id":"AAAAAKwhqVBA0-5Isxn7p1MwRCUBAAAABgAAAG4zAABNUT09AAAAAA==",
  "jobTitle":"Psychologist, sport and exercise",
  "userPrincipalName":"user1",
  "officeLocation":"Kopenhagen",
  "mail":"user1@kopano-dev.local",
  "givenName":"Jonas"
}%
```

```
./test/curl.sh https://kopano-dev.local/api/gc/v1/users/AAAAAKwhqVBA0-5Isxn7p1MwRCUBAAAABgAAAG4zAABNUT09AAAAAA==
{
  "mobilePhone":"05416 169130",
  "@odata.context":"\/api\/gc\/v1\/users\/AAAAAKwhqVBA0-5Isxn7p1MwRCUBAAAABgAAAG4zAABNUT09AAAAAA==",
  "surname":"Brekke",
  "displayName":"Jonas Brekke",
  "id":"AAAAAKwhqVBA0-5Isxn7p1MwRCUBAAAABgAAAG4zAABNUT09AAAAAA==",
  "jobTitle":"Psychologist, sport and exercise",
  "userPrincipalName":"user1",
  "officeLocation":"Kopenhagen",
  "mail":"user1@kopano-dev.local",
  "givenName":"Jonas"
}%
```

For further information and documentation on the supported endpoints, see
the [GRAPI](https://stash.kopano.io/projects/KC/repos/grapi) project README.

## Debugging

Sometimes it is useful to see the request payload data which is sent/received
by this plugin from the upstream Kopano MFR. This can be done by using `socat`
with the sockets.

Example:
```
socat -t100 -x -v UNIX-LISTEN:/run/kopano-grapi/rest0.sock,mode=777,reuseaddr,fork UNIX-CONNECT:/run/kopano-grapi/rest0.sock.orig
```

The above example assumes that Kopano GRAPI is started and a rest socket file
has been renamed to rest0.sock.orig so `socat` can forward the traffic to it.
