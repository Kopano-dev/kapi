# Kopanp API

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
./bin/kopano-apid serve --listen 127.0.0.1:8039 \
  --gc-socket-path=/run/kopano/fleet-runner
```

Where `--gc-socke-path` points to a folder containing files ending with `.sock`.
The folder is scanned for these files on startup and uses all found `.sock` files
as upstreams for the API proxy.

## Run unit tests

```
cd ~/go/src/stash.kopano.io/kc/konnect
make test
