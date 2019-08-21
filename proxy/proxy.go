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
	"net/http"
)

// A HTTPProxyHandler is a HTTP handler with response code and error cababilities.
type HTTPProxyHandler interface {
	ServeHTTP(rw http.ResponseWriter, req *http.Request) (int, error)
}

// The HTTPProxyHandlerFunc type is an adapter to allow the use of ordinary
// functions as HTTP oroxy handler. If f is with the appropriate signature,
// HTTPProxyHandlerFunc(f) is a HTTPProxyHandler that calls f.
type HTTPProxyHandlerFunc func(http.ResponseWriter, *http.Request) (int, error)

// ServeHTTP calls f(rw, req)
func (f HTTPProxyHandlerFunc) ServeHTTP(rw http.ResponseWriter, req *http.Request) (int, error) {
	return f(rw, req)
}
