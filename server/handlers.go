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
	"strings"

	kcoidc "stash.kopano.io/kc/libkcoidc"

	"stash.kopano.io/kc/kapi/auth"
	"stash.kopano.io/kc/kapi/proxy"
)

const (
	// AuthRequestHeaderName defines the request header which holds the ID of
	// the authenticated user.
	AuthRequestHeaderName = "X-Kopano-UserEntryID"
)

// HealthCheckHandler a http handler return 200 OK when server health is fine.
func (s *Server) HealthCheckHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

// AccessTokenRequired parses incoming bearer authentication and injects the
// subject of the token into the request as header.
func (s *Server) AccessTokenRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var err error

		var claims *kcoidc.ExtraClaimsWithType
		var authenticatedUserID string

		// TODO(longsleep): This code should be at a central location. It can
		// also be found in konnect.
		authHeader := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
		switch authHeader[0] {
		case "Bearer":
			if len(authHeader) != 2 {
				err = errors.New("invalid Bearer authorization header format")
				break
			}
			authenticatedUserID, _, claims, err = s.provider.ValidateTokenString(req.Context(), authHeader[1])

		default:
			err = errors.New("Bearer authorization required")
		}

		if err == nil {
			if claims != nil && claims.KCTokenType() == kcoidc.TokenTypeKCAccess {
				// TODO(longsleep): Support cases where the Subject is not a user entry ID.
				err = claims.Valid()
			} else {
				err = errors.New("missing access token claim")
			}
		}

		if err != nil {
			if s.allowAuthPassthrough {
				// NOTE(longsleep): Check for pass through of auth data.
				authenticatedUserID = req.Header.Get(AuthRequestHeaderName)
				if authenticatedUserID != "" {
					err = nil
				}
			}
		}

		if err == nil && authenticatedUserID != "" {
			req.Header.Set(AuthRequestHeaderName, authenticatedUserID)
			req = req.WithContext(auth.ContextWithAuthenticatedUserID(req.Context(), authenticatedUserID))
		} else {
			req.Header.Del(AuthRequestHeaderName)
		}

		if err != nil {
			s.logger.WithError(err).Debugln("access token required")
			http.Error(rw, "", http.StatusForbidden)
			return
		}

		next.ServeHTTP(rw, req)
	})
}

// HandleWithProxy returns a http handler to proxy requests to workers using the
// provided proxy.
func (s *Server) HandleWithProxy(proxy proxy.HTTPProxyHandler, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if proxy == nil {
			next.ServeHTTP(rw, req)
			return
		}

		status, err := proxy.ServeHTTP(rw, req)
		if err != nil {
			s.logger.WithError(err).Errorln("proxy request failed")
			http.Error(rw, "", status)
		}
	})
}
