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

	"stash.kopano.io/kc/kopano-api/plugins"
	"stash.kopano.io/kc/kopano-api/proxy"
)

// KopanoGroupwareCorePlugin implements the Kopano Groupware Core API within
// Kopano API.
type KopanoGroupwareCorePlugin struct {
	mutex  sync.RWMutex
	exitCh chan bool

	ctx context.Context
	srv plugins.ServerV1

	proxy proxy.Proxy
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
		return fmt.Errorf("KOPANOGC_REST_SOCKET value is invalid: %v", err)
	}

	if fp, err := os.Stat(socketPath); err != nil || !fp.IsDir() {
		return fmt.Errorf("KOPANO_GC_REST_SOCKETS does not exist or is not directory: %v", err)
	}

	// Start looking for sockets asynchronously to allow them to start later.
	go func() {
		err := p.initializeRest(ctx, socketPath)
		if err != nil {
			errCh <- err
		}
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
	handled := true
	switch path := req.URL.Path; {

	case strings.HasPrefix(path, "/api/gc/v0/"):
		p.srv.AccessTokenRequired(http.StripPrefix("/api/gc/v0/", http.HandlerFunc(p.handleRestV0))).ServeHTTP(rw, req)

	default:
		handled = false
	}

	return handled, nil
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
