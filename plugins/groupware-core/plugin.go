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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/cors"

	"stash.kopano.io/kc/kapi/plugins"
	"stash.kopano.io/kc/kapi/proxy"
	"stash.kopano.io/kc/kapi/version"
)

var pluginInfo = &plugins.InfoV1{
	ID:        "groupware-core",
	Version:   version.Version,
	BuildDate: version.BuildDate,
}

var scopesRequired = []string{"profile", "email", "kopano/gc"}
var apiV0Enabled = false

// KopanoGroupwareCorePlugin implements the Kopano Groupware Core API within
// Kopano API.
type KopanoGroupwareCorePlugin struct {
	mutex  sync.RWMutex
	exitCh chan bool

	ctx context.Context
	srv plugins.ServerV1

	cors *cors.Cors

	defaultProxy      proxy.HTTPProxyHandler
	subscriptionProxy proxy.HTTPProxyHandler
}

// Info returns the accociated plugins plugin.Info.
func (p *KopanoGroupwareCorePlugin) Info() *plugins.InfoV1 {
	return pluginInfo
}

// Initialize initizalizes the accociated plugin.
func (p *KopanoGroupwareCorePlugin) Initialize(ctx context.Context, errCh chan<- error, srv plugins.ServerV1) error {
	p.ctx = ctx
	p.srv = srv

	srv.Logger().Debugln("groupware-core: initialize")

	socketPath := os.Getenv("KOPANO_GC_REST_SOCKETS")
	if socketPath == "" {
		return fmt.Errorf("KOPANO_GC_REST_SOCKETS environment variable is not set but required")
	}

	socketPath, err := filepath.Abs(socketPath)
	if err != nil {
		return fmt.Errorf("KOPANO_GC_REST_SOCKETS value is invalid: %v", err)
	}

	if fp, err := os.Stat(socketPath); err != nil || !fp.IsDir() {
		return fmt.Errorf("KOPANO_GC_REST_SOCKETS does not exist or is not a directory: %v", err)
	}

	if os.Getenv("KOPANO_GC_REST_ALLOW_CORS") == "1" {
		p.srv.Logger().Warnln("groupware-core: CORS support enabled")
		p.cors = cors.AllowAll()
	}

	scopesRequiredString := os.Getenv("KOPANO_GC_REQUIRED_SCOPES")
	if scopesRequiredString != "" {
		scopesRequired = strings.Split(scopesRequiredString, " ")
	}
	p.srv.Logger().WithField("required_scopes", scopesRequired).Infoln("groupware-core: access requirements set up")

	if os.Getenv("KOPANO_GC_ENABLE_API_V0") == "1" {
		apiV0Enabled = true
		p.srv.Logger().Warnln("groupware-core: obsolete insecure API v0 endpoints enabled")
	}

	// Start looking for rest sockets asynchronously to allow them to start later.
	go func() {
		pr, err := p.initializeProxy(ctx, socketPath, "rest*.sock")
		if err != nil {
			errCh <- err
			return
		}
		if pr == nil {
			return
		}

		p.mutex.Lock()
		p.defaultProxy = pr
		p.mutex.Unlock()
		p.srv.Logger().Debugf("groupware-core: enabled default api proxy")
	}()

	// Start looking for subscriptions ockets asynchronously to allow them to start later.
	go func() {
		pr, err := p.initializeProxy(ctx, socketPath, "notify*.sock")
		if err != nil {
			errCh <- err
			return
		}
		if pr == nil {
			return
		}

		p.mutex.Lock()
		p.subscriptionProxy = pr
		p.mutex.Unlock()
		p.srv.Logger().Debugf("groupware-core: enabled subscription proxy")
	}()

	return nil
}

// Close closes the accociated plugin.
func (p *KopanoGroupwareCorePlugin) Close() error {
	p.srv.Logger().Debugln("groupware-core: close")

	close(p.exitCh)

	return nil
}

// ServeHTTP serves HTTP requests.
func (p *KopanoGroupwareCorePlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) (bool, error) {
	var handler http.Handler

	// Find handler.
	switch path := req.URL.Path; {
	case strings.HasPrefix(path, "/api/gc/v1/subscriptions"):
		handler = p.srv.AccessTokenRequired(http.HandlerFunc(p.handleSubscriptionsV1), scopesRequired)

	case apiV0Enabled && strings.HasPrefix(path, "/api/gc/v0/subscriptions"):
		// Backwards compatibility - rewrite URL to v1.
		req.URL.Path = strings.Replace(req.URL.Path, "/api/gc/v0/", "/api/gc/v1/", 1)
		handler = p.srv.AccessTokenRequired(http.HandlerFunc(p.handleSubscriptionsV1), nil)

	case strings.HasPrefix(path, "/api/gc/v1/"):
		handler = p.srv.AccessTokenRequired(http.HandlerFunc(p.handleDefaultV1), scopesRequired)

	case apiV0Enabled && strings.HasPrefix(path, "/api/gc/v0/"):
		// Backwards compatibility - rewrite URL to v1.
		req.URL.Path = strings.Replace(req.URL.Path, "/api/gc/v0/", "/api/gc/v1/", 1)
		handler = p.srv.AccessTokenRequired(http.HandlerFunc(p.handleDefaultV1), nil)
	}

	if handler == nil {
		// Fast exit.
		return false, nil
	}

	// Add support for CORS if configured.
	if p.cors != nil {
		handler = p.cors.Handler(handler)
	}

	// Execute handler.
	handler.ServeHTTP(rw, req)

	return true, nil
}

// Register is the exported registration entry point as loaded by Kopano API to
// register plugins.
var Register plugins.RegisterPluginV1 = func() plugins.PluginV1 {
	return &KopanoGroupwareCorePlugin{
		exitCh: make(chan bool, 1),
	}
}

// NOTE(longsleep): Keep main() to make the linter happy.
func main() {}
