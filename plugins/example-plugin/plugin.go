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

package main

import (
	"context"
	"errors"
	"net/http"

	"stash.kopano.io/kc/kapi/plugins"
	"stash.kopano.io/kc/kapi/version"
)

var pluginInfo = &plugins.InfoV1{
	ID:        "example-plugin",
	Version:   version.Version,
	BuildDate: version.BuildDate,
}

// ExamplePlugin implements an example plugin for Kopano API.
type ExamplePlugin struct {
	ctx context.Context
	srv plugins.ServerV1
}

// Info returns the accociated plugins plugin.Info.
func (p *ExamplePlugin) Info() *plugins.InfoV1 {
	return pluginInfo
}

// Initialize initizalizes the accociated plugin.
func (p *ExamplePlugin) Initialize(ctx context.Context, errCh chan<- error, srv plugins.ServerV1) error {
	p.ctx = ctx
	p.srv = srv

	srv.Logger().Debugln("example-plugin: initialize")
	return nil
}

// Close closes the accociated plugin.
func (p *ExamplePlugin) Close() error {
	p.srv.Logger().Debugln("example-plugin: close")

	return nil
}

// ServeHTTP serves HTTP requests.
func (p *ExamplePlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) (bool, error) {
	p.srv.Logger().Debugln("example-plugin: serveHTTP", req.URL)

	handled := true
	switch path := req.URL.Path; {
	case path == "/example-plugin/test":
		p.testHandler(rw, req)
	case path == "/example-plugin/auth-required/test":
		p.srv.AccessTokenRequired(http.HandlerFunc(p.testHandler), nil).ServeHTTP(rw, req)
	case path == "/example-plugin/fake-error":
		return false, errors.New("fake error")
	default:
		handled = false
	}

	return handled, nil
}

// Register is the exported registration entry point as loaded by Kopano API to
// register plugins.
var Register plugins.RegisterPluginV1 = func() plugins.PluginV1 {
	return &ExamplePlugin{}
}

// NOTE(longsleep): Keep main() to make the linter happy.
func main() {}
