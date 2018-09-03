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
	"os"
	"path/filepath"
	"time"

	"stash.kopano.io/kc/kapi/proxy"
	"stash.kopano.io/kc/kapi/proxy/httpproxy"
)

var restProxyConfiguration = &httpproxy.Configuration{
	Policy:      "least_conn",
	FailTimeout: 500 * time.Millisecond,
	MaxFails:    1,
	MaxConns:    0,
	Keepalive:   100,
	TryDuration: 1 * time.Second,
	TryInterval: 50 * time.Millisecond,
}

func (p *KopanoGroupwareCorePlugin) initializeProxy(ctx context.Context, socketPath string, pattern string) (proxy.HTTPProxyHandler, error) {
	p.srv.Logger().Debugf("grapi: looking for proxy %s files in %s", pattern, socketPath)

	var err error
	var init bool
	var count int
	for {
		for {
			if fp, statErr := os.Stat(socketPath); statErr != nil || !fp.IsDir() {
				err = statErr
				break
			}

			socketPaths, globErr := filepath.Glob(fmt.Sprintf("%s/%s", socketPath, pattern))
			if globErr != nil {
				err = globErr
				break
			}
			if len(socketPaths) == 0 {
				err = fmt.Errorf("no proxy %s files found in socket-path", pattern)
				break
			}

			if !init {
				// Do another loop to avoid missing socket files when they
				// are just in process of creation.
				init = true
				break
			}

			pr, proxyErr := httpproxy.New("grapi", socketPaths, restProxyConfiguration)
			if proxyErr != nil {
				return nil, proxyErr
			}

			p.srv.Logger().Debugf("grapi: found %d %s upstream proxy workers", len(socketPaths), pattern)
			return pr, nil
		}

		if err != nil && count == 5 {
			p.srv.Logger().WithError(err).Warnf("grapi: waiting for proxy %s files to appear", pattern)
		}
		count++
		if count > 60 {
			count = 0
		}

		select {
		case <-p.exitCh:
			return nil, nil
		case <-ctx.Done():
			return nil, nil
		case <-time.After(1 * time.Second):
			// retry.
		}
	}
}

func (p *KopanoGroupwareCorePlugin) handleDefaultV1(rw http.ResponseWriter, req *http.Request) {
	p.mutex.RLock()
	proxy := p.defaultProxy
	p.mutex.RUnlock()

	// Proxy all.
	p.srv.HandleWithProxy(proxy, http.HandlerFunc(p.handleNoProxy)).ServeHTTP(rw, req)
}

func (p *KopanoGroupwareCorePlugin) handleSubscriptionsV1(rw http.ResponseWriter, req *http.Request) {
	p.mutex.RLock()
	proxy := p.subscriptionProxy
	p.mutex.RUnlock()

	// Proxy all.
	p.srv.HandleWithProxy(proxy, http.HandlerFunc(p.handleNoProxy)).ServeHTTP(rw, req)
}

func (p *KopanoGroupwareCorePlugin) handleNoProxy(rw http.ResponseWriter, req *http.Request) {
	// NOTE(longsleep): This handler is only reached when no proxy is available.

	p.srv.Logger().WithError(errors.New("proxy not configured")).Errorln("grapi: proxy request not possible")
	http.Error(rw, "", http.StatusBadGateway)
}
