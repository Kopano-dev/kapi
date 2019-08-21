/*
 * Copyright 2019 Kopano and its licensors
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

package httpproxy

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"stash.kopano.io/kgol/rndm"

	"stash.kopano.io/kc/kapi/proxy"
)

type stickyProxyHandler struct {
	CookieName        string
	CookiePath        string
	HeaderName        string
	Nocache           bool
	SetUpstreamHeader bool
	SetCookie         bool
}

func newStickyProxyHandler(stickyRule string) (*stickyProxyHandler, error) {
	scanner := bufio.NewScanner(strings.NewReader(stickyRule))
	scanner.Split(bufio.ScanWords)

	sph := &stickyProxyHandler{
		CookiePath: "/",
	}

	for scanner.Scan() {
		t := scanner.Text()
		switch t {
		case "set-cookie":
			sph.SetCookie = true
			fallthrough

		case "cookie":
			scanner.Scan()
			sph.CookieName = scanner.Text()
			if sph.CookieName == "" {
				return nil, fmt.Errorf("sticky rule cookie is missing an argument")
			}

		case "set-upstream-header":
			scanner.Scan()
			sph.HeaderName = scanner.Text()
			if sph.HeaderName == "" {
				return nil, fmt.Errorf("sticky rule set-header is missing an argument")
			}
			sph.SetUpstreamHeader = true

		case "nocache":
			sph.Nocache = true

		case "cookie-path":
			scanner.Scan()
			sph.CookiePath = scanner.Text()
			if sph.CookiePath == "" {
				return nil, fmt.Errorf("sticky rule path is missing an argument")
			}

		default:
			return nil, fmt.Errorf("unknown sticky rule token: %v", t)
		}
	}

	return sph, nil
}

func (sph *stickyProxyHandler) Handler(next proxy.HTTPProxyHandler) proxy.HTTPProxyHandler {
	return proxy.HTTPProxyHandlerFunc(func(rw http.ResponseWriter, req *http.Request) (int, error) {
		// Prepare our way of operation.
		var setCookie bool
		var stickyValue string
		if sph.CookieName != "" {
			// If a cookie name is set, check if it was sent.
			cookie, _ := req.Cookie(sph.CookieName)
			if cookie != nil {
				stickyValue = cookie.Value
			} else if sph.SetCookie {
				// No sticky cookie, create new random value.
				stickyValue = rndm.GenerateRandomString(8)
				setCookie = true
			}
			if sph.SetUpstreamHeader && sph.HeaderName != "" && stickyValue != "" {
				// Set cookie value to header value. This makes it possible that
				// the next handler selects backend based on header value.
				req.Header.Set(sph.HeaderName, stickyValue)
			}
		}
		// Pre inject our stuff into the response headers.
		if sph.Nocache {
			// Add Cache-Control header if not already in response.
			if rw.Header().Get("Cache-Control") == "" {
				rw.Header().Set("Cache-Control", "nocache")
			}
		}
		if setCookie {
			// NOTE(longsleep): Cookie is set without domain.
			http.SetCookie(rw, &http.Cookie{
				Name:     sph.CookieName,
				Value:    stickyValue,
				Path:     sph.CookiePath,
				HttpOnly: false,
			})
		}
		// Execute.
		return next.ServeHTTP(rw, req)
	})
}
