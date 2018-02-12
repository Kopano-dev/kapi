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

package main

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
	v1.Handle("/webhook", p.srv.AccessTokenRequired(p.MakeHTTPWebhookRegisterHandler(v1))).
		Methods(http.MethodPost)
	v1.HandleFunc("/stream/websocket", p.HTTPWebsocketHandler).
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

// HTTPWebsocketHandler implements the HTTP handler for stream websocket requests.
func (p *PubsPlugin) HTTPWebsocketHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(rw, "", http.StatusMethodNotAllowed)
		return
	}

	err := p.handleWebsocketConnect(req.Context(), "", rw, req)
	if err != nil {
		p.srv.Logger().WithError(err).Errorln("pubs: stream websocket connection failed")
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}
}
