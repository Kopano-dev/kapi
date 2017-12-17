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
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"stash.kopano.io/kc/kopano-api/proxy/httpproxy"
)

var restProxyConfiguration = &httpproxy.Configuration{
	Policy:      "least_conn",
	FailTimeout: 20 * time.Millisecond,
	MaxFails:    1,
	MaxConns:    8,
	Keepalive:   100,
	TryDuration: 1 * time.Second,
	TryInterval: 50 * time.Millisecond,
}

func (p *KopanoGroupwareCorePlugin) initializeRest(ctx context.Context, socketPath string) error {
	p.srv.Logger().Debugf("groupware-core: looking for .sock files in %s", socketPath)

	var err error
	var init bool
	for {
		for {
			socketPaths, globErr := filepath.Glob(fmt.Sprintf("%s/*.sock", socketPath))
			if globErr != nil {
				err = globErr
				break
			}
			if len(socketPaths) == 0 {
				err = fmt.Errorf("no .sock files found in socket-path")
				break
			}

			if !init {
				// Do another loop to avoid missing socket files when they
				// are just in process of creation.
				init = true
				break
			}

			pr, proxyErr := httpproxy.New("groupware-core", socketPaths, restProxyConfiguration)
			if proxyErr != nil {
				return proxyErr
			}

			p.mutex.Lock()
			p.proxy = pr
			p.srv.Logger().Debugf("groupware-core: enabled proxy with %d upstream workers", len(socketPaths))
			p.mutex.Unlock()
			return nil
		}

		if err != nil {
			p.srv.Logger().WithError(err).Warnln("groupware-core: waiting for .sock files to appear")
		}

		select {
		case <-p.exitCh:
			return nil
		case <-ctx.Done():
			return nil
		case <-time.After(1 * time.Second):
			// retry.
		}
	}
}

func (p *KopanoGroupwareCorePlugin) handleRestV0(rw http.ResponseWriter, req *http.Request) {
	p.mutex.RLock()
	proxy := p.proxy
	p.mutex.RUnlock()

	// Proxy all.
	p.srv.HandleWithProxy(proxy, http.HandlerFunc(p.handleNoProxy)).ServeHTTP(rw, req)
}

func (p *KopanoGroupwareCorePlugin) handleNoProxy(rw http.ResponseWriter, req *http.Request) {
	// NOTE(longsleep): This handler is only reached when no proxy is available.

	p.srv.Logger().WithError(errors.New("proxy not configured")).Errorln("groupware-core: proxy request not possible")
	http.Error(rw, "", http.StatusBadGateway)
}
