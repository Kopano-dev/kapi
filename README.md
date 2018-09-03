# Kopano API

Kopano API provides a web service with the endpoints to interface with Kopano
via HTTP APIs.

## TL;DW

Make sure you have Go 1.8 or later installed. This assumes your GOPATH is `~/go` and
you have `~/go/bin` in your $PATH and you have [Glide](https://github.com/Masterminds/glide)
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

Kopano API supports plugins. Plugins can be used to extend HTTP routes served
by Kopano API. A example plugin can be found in `plugins/example-plugin`.

### grapi: Kopano Groupware REST plugin

Kopano API includes the plugin for Kopano Groupware REST. This plugin provides
access to Kopano Groupware RESTful API via `/api/gc/` URL routing prefix.

To specify where the Grapi plugin can find its required backend sockets, specify
the `KOPANO_GRAPI_SOCKETS` environment variable to point to the base directory
location. All `rest*.sock` files in that directory will be used as upstream
proxy paths for the REST api and all `notify*.sock` files in that directory will
be used as upstream proxy paths for the subscription socket API.

### pubs: Kopano Pubsub and Webhook plugin

Kopano API includes a pub/sub system and webhooko system via the Pubs plugin. To
specify the cryptographic secret for the Pubs plugin use the environment
variable `KOPANO_PUBS_SECRET_KEY`. For more information on the Pubs plugin look
at 'plugins/pubs/README.md'.

## Run unit tests

```
cd ~/go/src/stash.kopano.io/kc/kapi
make test
```
