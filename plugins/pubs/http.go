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

package pubs

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	httpBaseURL              = "/api/pubs/v1/"
	webhookRouterIdentifier  = "webhook-by-id-and-token"
	websocketRouteIdentifier = "stream-websocket"
)

func (p *PubsPlugin) addRoutes(ctx context.Context, router *mux.Router) http.Handler {
	v1 := router.PathPrefix(httpBaseURL).Subrouter()

	v1.Handle(
		"/webhook/{id}/{token}/{envelope}", p.MakeHTTPWebhookPublishHandler(v1)).
		Methods(http.MethodPost)
	v1.Handle("/webhook/{id}/{token}", p.MakeHTTPWebhookPublishHandler(v1)).
		Methods(http.MethodPost).
		Name(webhookRouterIdentifier)
	v1.Handle("/webhook", p.srv.AccessTokenRequired(p.MakeHTTPWebhookRegisterHandler(v1), scopesRequired)).
		Methods(http.MethodPost)
	v1.Handle("/stream/connect", p.srv.AccessTokenRequired(p.MakeHTTPWebsocketConnectHandler(v1), scopesRequired))
	v1.HandleFunc("/stream/websocket/{key}", p.HTTPWebsocketHandler).
		Methods(http.MethodGet).
		Name(websocketRouteIdentifier)

	return router
}

// ServeHTTP serves HTTP requests.
func (p *PubsPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) (bool, error) {
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

// MakeHTTPWebhookRegisterHandler implements the HTTP handler for registering webhooks.
func (p *PubsPlugin) MakeHTTPWebhookRegisterHandler(router *mux.Router) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := p.handleWebhookRegister(req.Context(), router, rw, req)
		if err != nil {
			p.srv.Logger().WithError(err).Errorln("pubs: webhook register failed")
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}
	})
}

// MakeHTTPWebhookPublishHandler creates the HTTP handler for registering webhooks.
func (p *PubsPlugin) MakeHTTPWebhookPublishHandler(router *mux.Router) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := p.handleWebhookPublish(req.Context(), rw, req)
		if err != nil {
			p.srv.Logger().WithError(err).Errorln("pubs: webhook publish failed")
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}
	})
}

// MakeHTTPWebsocketConnectHandler createss the HTTP handler for rtm.connect.
func (p *PubsPlugin) MakeHTTPWebsocketConnectHandler(router *mux.Router) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// create random URL to websocket endpoint
		key, err := p.handleWebsocketConnect(req.Context())
		if err != nil {
			p.srv.Logger().WithError(err).Errorln("pubs: stream websocket connect failed")
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}

		route := router.Get(websocketRouteIdentifier)
		websocketURI, err := route.URLPath("key", key)
		if err != nil {
			p.srv.Logger().WithError(err).Errorln("pubs: stream websocket connect url generation failed")
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}

		response := &streamWebsocketConnectResponse{
			StreamURL: websocketURI.String(),
		}

		err = WriteJSON(rw, http.StatusOK, response, "")
		if err != nil {
			p.srv.Logger().WithError(err).Errorln("pubs: failed to write JSON response")
		}
	})
}

// HTTPWebsocketHandler implements the HTTP handler for stream websocket requests.
func (p *PubsPlugin) HTTPWebsocketHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(rw, "", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(req)
	key, ok := vars["key"]
	if !ok {
		http.NotFound(rw, req)
		return
	}

	err := p.handleWebsocketConnection(req.Context(), key, rw, req)
	if err != nil {
		p.srv.Logger().WithError(err).Errorln("pubs: stream websocket connection failed")
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}
}
