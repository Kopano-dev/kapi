module stash.kopano.io/kc/kapi

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/caddyserver/caddy v1.0.5
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cskr/pubsub v1.0.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang-migrate/migrate v3.5.4+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.1
	github.com/klauspost/cpuid v1.2.3 // indirect
	github.com/longsleep/go-metrics v0.0.0-20191013204616-cddea569b0ea
	github.com/mattn/go-sqlite3 v1.13.0
	github.com/miekg/dns v1.1.27 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/procfs v0.0.10 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.6
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/mcuadros/go-syslog.v2 v2.3.0 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	stash.kopano.io/kc/libkcoidc v0.8.1
	stash.kopano.io/kgol/oidc-go v0.3.1 // indirect
	stash.kopano.io/kgol/rndm v1.1.0
	stash.kopano.io/kwm/kwmserver v1.1.0
)

replace github.com/lucas-clemente/quic-go => github.com/lucas-clemente/quic-go v0.7.1-0.20200818072714-1d8e4c08910c
