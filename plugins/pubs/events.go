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
	"encoding/json"

	"stash.kopano.io/kwm/kwmserver/signaling/connection"
)

// OnConnect is called for new connections.
func (p *PubsPlugin) OnConnect(c *connection.Connection) error {
	c.Logger().Debugln("pubs: stream websocket connect")

	return p.onSubInit(c)
}

// OnDisconnect is called after a connection has closed.
func (p *PubsPlugin) OnDisconnect(c *connection.Connection) error {
	c.Logger().Debugln("pubs: stream websocket disconnect")

	return p.onUnsubAll(c)
}

// OnBeforeDisconnect is called before a connection is closed. An indication why
// the connection will be closed is provided with the passed error.
func (p *PubsPlugin) OnBeforeDisconnect(c *connection.Connection, err error) error {
	//p.srv.Logger().Debugln("pubs: stream websocket OnBeforeDisconnect")

	return nil
}

// OnText is called when the provided connection received a text message. The
// message payload is provided as []byte in the msg parameter.
func (p *PubsPlugin) OnText(c *connection.Connection, msg []byte) error {
	//p.srv.Logger().Debugf("websocket OnText: %s", msg)

	var envelope streamEnvelope
	err := json.Unmarshal(msg, &envelope)
	if err != nil {
		return err
	}

	err = nil
	switch envelope.Type {
	case streamEnvelopeTypeNameSub:
		fallthrough
	case streamEnvelopeTypeNameUnsub:
		fallthrough
	case streamEnvelopeTypeNamePub:
		err = p.onPubSub(c, &envelope)

	default:
		c.Logger().WithField("type", envelope.Type).Warnln("pubs: unknown incoming message type")
		return nil
	}

	if err != nil {
		return err
	}

	if envelope.State != "" {
		// Client state received, reply with ack.
		var event []byte
		event, err = PrettyJSON(&streamEnvelope{
			Type:  streamEnvelopeTypeAck,
			State: envelope.State,
		})
		if err == nil {
			err = c.RawSend(event)
		}
	}

	return err
}

// OnError is called, when the provided connection has encountered an error. The
// provided error is the error encountered. Any return value other than nil,
// will result in a close of the connection.
func (p *PubsPlugin) OnError(c *connection.Connection, err error) error {
	return err
}
