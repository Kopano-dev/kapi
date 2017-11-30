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

package server

import (
	"errors"
	"net/http"
)

// HealthCheckHandler a http handler return 200 OK when server health is fine.
func (s *Server) HealthCheckHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

// ProxyHandler is a http handler to proxy requests to workers using the
// accociated proxy.
func (s *Server) ProxyHandler(rw http.ResponseWriter, req *http.Request) {
	s.mutex.RLock()
	proxy := s.proxy
	s.mutex.RUnlock()

	if proxy == nil {
		s.logger.WithError(errors.New("proxy not configured")).Errorln("proxy request not possible")
		http.Error(rw, "", http.StatusBadGateway)
		return
	}

	status, err := proxy.ServeHTTP(rw, req)
	if err != nil {
		s.logger.WithError(err).Errorln("proxy request failed")
		http.Error(rw, "", status)
	}
}
