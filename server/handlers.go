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

	"github.com/dgrijalva/jwt-go"
	"stash.kopano.io/kc/konnect"
	"stash.kopano.io/kc/konnect/oidc"

	"stash.kopano.io/kc/kopano-api/proxy"
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
		var claims *konnect.AccessTokenClaims

		// TODO(longsleep): This code should be at a central location. It can
		// also be found in konnect.
		auth := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
		switch auth[0] {
		case oidc.TokenTypeBearer:
			if len(auth) != 2 {
				err = oidc.NewOAuth2Error(oidc.ErrorOAuth2InvalidRequest, "Invalid Bearer authorization header format")
				break
			}
			claims = &konnect.AccessTokenClaims{}
			_, err = jwt.ParseWithClaims(auth[1], claims, func(token *jwt.Token) (interface{}, error) {
				// TODO(longsleep): validate!

				return nil, errors.New("validate of tokens not implemented")
			})
			err = nil //XXX(longsleep): Remove me once validation is implemented.
			if err == nil {
				// TODO(longsleep): Validate all claims.
				err = claims.Valid()
			}
			if err != nil {
				// Wrap as OAuth2 error.
				err = oidc.NewOAuth2Error(oidc.ErrorOAuth2InvalidToken, err.Error())
			}

		default:
			err = oidc.NewOAuth2Error(oidc.ErrorOAuth2InvalidRequest, "Bearer authorization required")
		}

		if claims != nil && claims.IsAccessToken {
			// Use sub
			req.Header.Set("X-Kopano-UserEntryID", claims.Subject)
			//s.logger.WithField("UserEntryID", req.Header.Get("X-Kopano-UserEntryID")).Debugln("proxy with auth")
		} else {
			err = errors.New("missing access token claim")
		}

		if err != nil && req.Header.Get("X-Kopano-UserEntryID") != "" {
			// XXX(longsleep): This allows insecure pass through of auth data. Remove this!
			err = nil
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
