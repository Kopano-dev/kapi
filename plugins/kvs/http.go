/*
 * Copyright 2018 Kopano and its licensors
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

package plugin

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	httpBaseURL = "/api/kvs/v1/"
)

func (p *KVSPlugin) addRoutes(ctx context.Context, router *mux.Router) http.Handler {
	v1 := router.PathPrefix(httpBaseURL).Subrouter()

	v1.PathPrefix("/kv/user/").Handler(http.StripPrefix(httpBaseURL+"kv/user/", p.srv.AccessTokenRequired(p.MakeHTTPUserKVHandler(v1), scopesRequired)))

	return router
}

// ServeHTTP serves HTTP requests.
func (p *KVSPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) (bool, error) {
	var handler http.Handler

	switch path := req.URL.Path; {
	case strings.HasPrefix(path, httpBaseURL):
		handler = p.handler
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

// MakeHTTPUserKVHandler creates the HTTP handler for the per user kv store.
func (p *KVSPlugin) MakeHTTPUserKVHandler(router *mux.Router) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		p.handleUserKV(rw, req)
	})
}
