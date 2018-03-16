/*
 * Copyright 2017 Kopano and its licensors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package httpproxy

import (
	"bytes"
	"text/template"
	"time"

	"github.com/mholt/caddy/caddyfile"
	caddyproxy "github.com/mholt/caddy/caddyhttp/proxy"

	"stash.kopano.io/kc/kapi/proxy"
)

var configurationTemplate = template.Must(template.New("Caddyfile").Parse(`
	proxy / {
		transparent

		policy {{.C.Policy}}
		fail_timeout {{.C.FailTimeout}}
		max_fails {{.C.MaxFails}}
		max_conns {{.C.MaxConns}}
		keepalive {{.C.Keepalive}}
		try_duration {{.C.TryDuration}}
		try_interval {{.C.TryInterval}}

		{{range .UpstreamURIs}}
		upstream unix://{{.}}
		{{end}}

		{{range .C.Extra}}
		{{.}}
		{{end}}
	}
`))

// Configuration defines configuration settings for a proxy.
type Configuration struct {
	Policy      string
	FailTimeout time.Duration
	MaxFails    uint
	MaxConns    uint
	Keepalive   uint
	TryDuration time.Duration
	TryInterval time.Duration
	Extra       []string
}

// DefaultConfiguration is the proxy configuration which is used by default.
var DefaultConfiguration = &Configuration{
	Policy:      "random",
	FailTimeout: 0,
	MaxFails:    1,
	MaxConns:    0,
	Keepalive:   8,
	TryDuration: 0,
	TryInterval: time.Duration(250) * time.Millisecond,
}

type configurationWithUpstreams struct {
	UpstreamURIs []string
	C            *Configuration
}

// New creates a new proxy identified by the provided name to the provided
// upstreamURIs..
func New(name string, upstreamURIs []string, configuration *Configuration) (proxy.HTTPProxyHandler, error) {
	if configuration == nil {
		configuration = DefaultConfiguration
	}

	var buf bytes.Buffer
	err := configurationTemplate.Execute(&buf, &configurationWithUpstreams{
		UpstreamURIs: upstreamURIs,
		C:            configuration,
	})
	if err != nil {
		return nil, err
	}

	// Setup proxy stuff, by creating a caddy file.
	dispenser := caddyfile.NewDispenser(name, &buf)
	upstreams, err := caddyproxy.NewStaticUpstreams(dispenser, "")
	if err != nil {
		return nil, err
	}

	proxy := &caddyproxy.Proxy{
		Next:      nil,
		Upstreams: upstreams,
	}

	return proxy, nil
}
