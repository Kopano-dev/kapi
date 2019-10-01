# Kopano API

Kopano API provides a web service with the endpoints to interface with Kopano
via HTTP APIs. The availablity of APIS is controlled by plugins. See the [Plugins](#Plugins)
section below for details.

# Technologies

- Go

## Build dependencies

Make sure you have Go 1.8 or later installed. This assumes your GOPATH is `~/go` and
you have `~/go/bin` in your $PATH and you have [Dep](https://golang.github.io/dep/)
installed as well.

## Building from source

```
mkdir -p ~/go/src/stash.kopano.io/kc
cd ~/go/src/stash.kopano.io/kc
git clone <THIS-PROJECT> kapi
cd kopano-api
make
```

## Running Kopano API

```
KOPANO_GRAPI_SOCKETS=/run/kopano-grapi ./bin/kapid serve \
  --listen 127.0.0.1:8039 \
  --plugins-path=./plugins \
  --plugins=grapi \
  --iss=https://mykonnect.local
```

Where `--plugins-path` points to a folder containing Kopano API plugin modules.
Add environment variables as needed by those plugins. See next chapter for
more information about plugins.

The `--plugins` parameter can be used to select what plugins should be enabled.
It takes a comma seperated value of plugin IDs as the plugin defined it during
its build time. If the `--plugins` parameter is empty (the default), all plugins
found will be activated.

The `--iss` parameter points to an OpenID Connect issuer with support for
discovery (Kopano Konnect). On start, the service will try discover OIDC details
and allow Bearer authentication with access tokens once successful. The `--iss`
parameter is mandatory.

## Plugins

Kopano API supports plugins to its behavior and ships with a bunch of
plugins to provide API endpoints from various data sources and different
purposes. An example plugin can be found in `plugins/example-plugin`.

### grapi: Kopano Groupware REST plugin (GRAPI)

Kopano API includes the plugin for Kopano Groupware REST. This plugin provides
access to Kopano Groupware RESTful API via `/api/gc/` URL routing prefix.

To specify where the grapi plugin can find its required GRAPI backend sockets,
specify the `KOPANO_GRAPI_SOCKETS` environment variable to point to the base
directory location. All `rest*.sock` files in that directory will be used as
upstream proxy paths for the REST api and all `notify*.sock` files in that
directory will be used as upstream proxy paths for the subscription socket API.

See the [grapi plugin README](https://stash.kopano.io/projects/KC/repos/kapi/browse/plugins/grapi/README.md) for further details.

### pubs: Kopano Pubsub and Webhook plugin

Kopano API includes a pub/sub system and webhook system via the Pubs plugin,
routed to the `/api/pubs` URL routing prefix. To specify the cryptographic
secret for the Pubs plugin use the environment variable
`KOPANO_PUBS_SECRET_KEY`. For more information on the Pubs plugin look at
'plugins/pubs/README.md'.

See the [pubs plugin README](https://stash.kopano.io/projects/KC/repos/kapi/browse/plugins/pubs/README.md) for further details.

### kvs: Kopano Key Value Store plugin

Kopano API inclues a key value store via the kvs plugin, routed to the
`/api/kvs` URL routing prefix. Kvs plugin needs configuration for its persistent
storage layer. Look at 'plugins/kvs/README.md' for more information.

See the [kvs plugin README](https://stash.kopano.io/projects/KC/repos/kapi/browse/plugins/kvs/README.md) for further details.

## Run unit tests

```
cd ~/go/src/stash.kopano.io/kc/kapi
make test
```

## Testing the Kopano API

To test, some prerequisites are needed. A full fledged setup with TLS web server,
authentication provider and backend is strongly suggested. An quick way to set
this up is using a [Kopano Docker environment](https://github.com/kopano-dev/kopano-docker) which
provides:

  - TLS web server with [Kopano Web](https://stash.kopano.io/projects/KGOL/repos/kweb),
  - Authentication with [Kopanp Konnect](https://stash.kopano.io/projects/KC/repos/konnect)
  - Backend with [GRAPI](https://stash.kopano.io/projects/KC/repos/grapi) and [Kopano Groupware](https://stash.kopano.io/projects/KC/repos/kopanocore)

Once you have all the bits in place and set up correctly, look at the `test`
folder in this project for a bunch of scripts and helpers to simplify testing
and give you ideas how to access the APIs provided by kapi.
