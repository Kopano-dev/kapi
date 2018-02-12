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
)

const (
	streamEnvelopeTypeNameSub    = "sub"
	streamEnvelopeTypeNameUnsub  = "unsub"
	streamEnvelopeTypeCloseTopic = "closeTopic"
	streamEnvelopeTypeNamePub    = "pub"

	streamEnvelopeTypeHello = "hello"
	streamEnvelopeTypeAck   = "ack"
	streamEnvelopeTypeEvent = "event"
)

type webhookPubTokenData struct {
	ID    string `json:"id"`
	Topic string `json:"topic"`
}

type webhookRegisterResponse struct {
	ID     string `json:"id"`
	Topic  string `json:"topic"`
	PubURL string `json:"pubUrl"`
}

type streamEnvelope struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data,omitempty"`
	Info  json.RawMessage `json:"info,omitempty"`
	State string          `json:"state,omitempty"`
}

type streamTopicDefinition struct {
	Ref    string   `json:"ref,omitempty"`
	Topics []string `json:"topics,omitempty"`
}
