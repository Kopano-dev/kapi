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
git clone <THIS-PROJECT> kopano-api
cd kopano-api
make
```

## Running Kopano API

```
KOPANO_GC_REST_SOCKETS=/run/kopano/fleet-runner ./bin/kopano-apid serve \
  --listen 127.0.0.1:8039 \
  --plugins-path=./plugins \
  --plugins=groupware-core
```

Where `--plugins-path` points to a folder containing Kopano API plugin modules.
Add environment variables as needed by those plugins. See next chapter for
more information about plugins.

The `--plugins` parameter can be used to select what plugins should be enabled.
It takes a comma seperated value of plugin IDs as the plugin defined it during
its build time. If the `--plugins` parameter is empty (the default), all plugins
found will be activated.

## Plugins

Kopano API supports plugins. Plugins can be used to extend HTTP routes served
by Kopano API. A example plugin can be found in `plugins/example-plugin`.

### Kopano Groupware Core plugin

Kopano API includes the plugin for Kopano Groupware Core. This plugin provides
access to Kopano Groupware Core RESTful API via `/api/gc/` URL routing prefix.

To specify where the Groupware Core plugin can find its required backend sockets
specify the `KOPANO_GC_REST_SOCKETS` environment variable to point to the base
directory location. All `.sock` files in that directory will be used as upstream
proxy paths.

## Run unit tests

```
cd ~/go/src/stash.kopano.io/kc/kopano-api
make test
```
