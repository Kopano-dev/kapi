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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"stash.kopano.io/kgol/rndm"
	"stash.kopano.io/kwm/kwmserver/signaling/connection"

	"stash.kopano.io/kc/kapi/auth"
)

// Buffer sizes for HTTP webhook requests.
const (
	maxRequestSize = 1024 * 5
)

func (p *PubsPlugin) handleWebhookRegister(ctx context.Context, router *mux.Router, rw http.ResponseWriter, req *http.Request) error {
	req.ParseForm()
	topic := req.Form.Get("topic")
	if topic == "" {
		topic = rndm.GenerateRandomString(32)
	}

	id := rndm.GenerateRandomString(32)
	tokenData := &webhookPubTokenData{
		ID:    id,
		Topic: topic,
	}
	token, err := p.cookie.Encode("pubs-webhook", tokenData)
	if err != nil {
		return err
	}

	route := router.Get(webhookRouterIdentifier)
	pubURL, err := route.URLPath("id", id, "token", token)
	if err != nil {
		return err
	}

	response := &webhookRegisterResponse{
		ID:     id,
		Topic:  topic,
		PubURL: pubURL.String(),
	}

	err = WriteJSON(rw, http.StatusOK, response, "")
	if err != nil {
		p.srv.Logger().WithError(err).Errorln("pubs: failed to write JSON response")
		return nil
	}

	p.srv.Logger().WithFields(logrus.Fields{
		"topic": topic,
		"id":    id,
	}).Debugln("pubs: registered webhook")

	return nil
}

func (p *PubsPlugin) handleWebhookPublish(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	token, ok := vars["token"]
	if !ok {
		http.NotFound(rw, req)
		return nil
	}

	// Decode token.
	tokenData := &webhookPubTokenData{}
	err := p.cookie.Decode("pubs-webhook", token, tokenData)
	if err != nil {
		p.srv.Logger().WithError(err).Debugln("pubs: failed to decode webhook publish token")
		http.Error(rw, "", http.StatusUnprocessableEntity)
		return nil
	}

	// TODO(longsleep): Add check if the topic in the token still exists. If not
	// return httpStatusUnprocessableEntity to let the caller know that the
	// topic went away and it should stop calling.

	req.ParseForm()
	validationToken := req.Form.Get("validationToken")
	if validationToken != "" {
		p.srv.Logger().WithFields(logrus.Fields{
			"id": tokenData.ID,
		}).Debugln("pubs: webhook incoming publish validation")

		// Simple validation support via a validationToken handshake request.
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
		io.WriteString(rw, validationToken)
		return nil
	}

	// Read request data, up to a maximum.
	msg, err := ioutil.ReadAll(io.LimitReader(req.Body, maxRequestSize))
	if err != nil {
		p.srv.Logger().WithError(err).WithField("id", tokenData.ID).Warnln("pubs: webhook publish size limit exceeded")
		http.Error(rw, "", http.StatusBadRequest)
		return nil
	}

	// p.srv.Logger().WithField("topic", tokenData.Topic).Debugf("pubs: webhook data received %s", msg)

	info, err := PrettyJSON(&streamTopicDefinition{
		Ref:    tokenData.ID,
		Topics: []string{tokenData.Topic},
	})
	if err != nil {
		return err
	}

	envelope := vars["envelope"]
	if envelope != "" {
		// Add JSON envelope.
		// FIXME(longsleep): This can create invalid JSON based on the provided data.
		msg = []byte(fmt.Sprintf("{\"type\":\"%s\",\"data\":%s}", envelope, msg))
	}

	// Marshal all to JSON.
	event, err := PrettyJSON(&streamEnvelope{
		Type: streamEnvelopeTypeEvent,
		Data: msg,
		Info: info,
	})
	if err != nil {
		// Return a bad request when stuff cannot be marshaled as JSON as this usually
		// means that the JSON payload received from the webhook request is invalid.
		p.srv.Logger().WithError(err).WithField("id", tokenData.ID).Warnln("pubs: webhook publish failed to marshal")
		http.Error(rw, "", http.StatusBadRequest)
		return nil
	}

	p.srv.Logger().WithFields(logrus.Fields{
		"id":   tokenData.ID,
		"size": len(msg),
	}).Debugln("pubs: webhook incoming publish data")

	p.pubsub.Pub(event, tokenData.Topic)

	rw.WriteHeader(http.StatusNoContent)

	return nil
}

func (p *PubsPlugin) handleWebsocketConnect(ctx context.Context) (string, error) {
	key := rndm.GenerateRandomString(connectKeySize)

	// Add key to table.
	record := &keyRecord{
		when: time.Now(),
	}

	authenticatedUserID, _ := auth.AuthenticatedUserIDFromContext(ctx)
	if authenticatedUserID == "" {
		return "", fmt.Errorf("request is not authorized")
	}

	record.user = &userRecord{
		id: authenticatedUserID,
	}

	p.keys.Set(key, record)

	p.srv.Logger().WithFields(logrus.Fields{
		"key": key,
	}).Debugln("pubs: registered websocket")

	return key, nil
}

func (p *PubsPlugin) handleWebsocketConnection(ctx context.Context, key string, rw http.ResponseWriter, req *http.Request) error {
	record, ok := p.keys.Pop(key)
	if !ok {
		http.NotFound(rw, req)
		return nil
	}

	kr := record.(*keyRecord)
	if kr.user == nil || kr.user.id == "" {
		http.Error(rw, "", http.StatusForbidden)
		return nil
	}

	ws, err := p.upgrader.Upgrade(rw, req, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		p.srv.Logger().WithError(err).Debugln("pubs: stream websocket handshake error")
		return nil
	} else if err != nil {
		return err
	}

	id := strconv.FormatUint(atomic.AddUint64(&p.count, 1), 10)

	loggerFields := logrus.Fields{
		"websocket_connection": id,
	}

	c, err := connection.New(ctx, ws, p, p.srv.Logger().WithFields(loggerFields), id)
	if err != nil {
		return err
	}

	c.Logger().WithFields(logrus.Fields{
		"key": key,
	}).Debugln("pubs: stream websocket incoming connection")

	go p.serveWebsocketConnection(c, id)

	return nil
}

func (p *PubsPlugin) serveWebsocketConnection(c *connection.Connection, id string) {
	p.connections.Set(id, c)
	c.ServeWS(p.ctx)
	p.connections.Remove(id)
}
