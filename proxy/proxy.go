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

package proxy

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/mholt/caddy/caddyfile"
	"github.com/mholt/caddy/caddyhttp/proxy"
)

var proxyConfiguration = template.Must(template.New("Caddyfile").Parse(`
	proxy / {
		transparent

		policy least_conn
		fail_timeout 20ms
		max_fails 1
		max_conns 8
		keepalive 100
		try_duration 1s
		try_interval 50ms

		{{range .}}
		upstream unix://{{.}}
		{{end}}
	}
`))

// A Proxy is a HTTP handler with response code and error cababilities.
type Proxy interface {
	ServeHTTP(rw http.ResponseWriter, req *http.Request) (int, error)
}

// New creates a new proxy to the provided socket paths.
func New(socketPaths []string) (Proxy, error) {
	var buf bytes.Buffer
	err := proxyConfiguration.Execute(&buf, socketPaths)
	if err != nil {
		return nil, err
	}

	// Setup proxy stuff, by creating a caddy file.
	dispenser := caddyfile.NewDispenser("filename", &buf)
	upstreams, err := proxy.NewStaticUpstreams(dispenser, "")
	if err != nil {
		return nil, err
	}

	proxy := &proxy.Proxy{
		Next:      nil,
		Upstreams: upstreams,
	}

	return proxy, nil
}
