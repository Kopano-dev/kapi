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
	"errors"

	"github.com/sirupsen/logrus"
	"stash.kopano.io/kgol/rndm"
	"stash.kopano.io/kwm/kwmserver/signaling/api-v1/connection"
)

type pubsubBinder struct {
	id string
	ch chan interface{}
}

func (p *PubsPlugin) onSubInit(c *connection.Connection) error {
	ch := p.pubsub.Sub(p.broadcast)
	binder := &pubsubBinder{
		id: rndm.GenerateRandomString(32),
		ch: ch,
	}
	c.Bind(binder)

	// Say hello.
	info, err := PrettyJSON(&streamTopicDefinition{
		Ref: binder.id,
	})
	if err != nil {
		return err
	}
	event, err := PrettyJSON(&streamEnvelope{
		Type: streamEnvelopeTypeHello,
		Info: info,
	})
	if err != nil {
		return nil
	}
	err = c.RawSend(event)
	if err != nil {
		c.Logger().WithError(err).Warnf("pubs: error while sending hello to connection %s", c.ID())
		return err
	}

	// Initialize forwarder.
	go func() {
		defer func() {
			c.Logger().WithField("id", binder.id).Debugln("pubs: pubsub connection has ended")
		}()
		var err error
		for {
			select {
			case payload := <-ch:
				if payload == nil {
					// Normal close, do nothing.
					return
				}
				err = c.RawSend(payload.([]byte))
				if err != nil {
					c.Logger().WithError(err).WithField("id", binder.id).Warnln("pubs: error while sending to connection")
					// Close connection if it is not able to get our pubsub data.
					c.Close()
					return
				}
			}
		}
	}()

	c.Logger().WithField("id", binder.id).Debugln("pubs: pubsub connection initialized")

	return nil
}

func (p *PubsPlugin) onPubSub(c *connection.Connection, envelope *streamEnvelope) error {
	var err error
	var topicDefinition streamTopicDefinition
	if len(envelope.Info) > 0 {
		err = json.Unmarshal(envelope.Info, &topicDefinition)
		if err != nil {
			return err
		}
	}

	switch envelope.Type {
	case streamEnvelopeTypeNameSub:
		err = p.onSub(c, &topicDefinition)

	case streamEnvelopeTypeNameUnsub:
		err = p.onUnsub(c, &topicDefinition)

	case streamEnvelopeTypeNamePub:
		err = p.onPub(c, &topicDefinition, envelope.Data)

	default:
		err = errors.New("unknown_pubsub_type")
	}

	return err
}

func (p *PubsPlugin) onSub(c *connection.Connection, topicDefinition *streamTopicDefinition) error {
	if len(topicDefinition.Topics) == 0 {
		return errors.New("pubsub_without_topics")
	}
	binder := c.Bound().(*pubsubBinder)
	if binder == nil {
		return errors.New("pubsub_not_bound")
	}

	c.Logger().WithFields(logrus.Fields{
		"topics": topicDefinition.Topics,
		"id":     binder.id,
	}).Debugln("pubs: sub with connection")
	p.pubsub.AddSub(binder.ch, topicDefinition.Topics...)

	return nil
}

func (p *PubsPlugin) onUnsubAll(c *connection.Connection) error {
	binder := c.Bound().(*pubsubBinder)
	if binder == nil {
		return errors.New("pubsub_not_bound")
	}

	c.Logger().WithField("id", binder.id).Debugln("pubs: unsub all with connection")
	p.pubsub.Unsub(binder.ch)
	return nil
}

func (p *PubsPlugin) onUnsub(c *connection.Connection, topicDefinition *streamTopicDefinition) error {
	if len(topicDefinition.Topics) == 0 {
		return errors.New("pubsub_without_topics")
	}
	binder := c.Bound().(*pubsubBinder)
	if binder == nil {
		return errors.New("pubsub_not_bound")
	}

	c.Logger().WithFields(logrus.Fields{
		"topics": topicDefinition.Topics,
		"id":     binder.id,
	}).Debugln("pubs: unsub with connection")
	p.pubsub.Unsub(binder.ch, topicDefinition.Topics...)

	return nil
}

func (p *PubsPlugin) onPub(c *connection.Connection, topicDefinition *streamTopicDefinition, msg []byte) error {
	if len(topicDefinition.Topics) == 0 {
		return errors.New("pubsub_without_topics")
	}
	binder := c.Bound().(*pubsubBinder)
	if binder == nil {
		return errors.New("pubsub_not_bound")
	}
	topicDefinition.Ref = binder.id

	info, err := PrettyJSON(topicDefinition)
	if err != nil {
		return err
	}

	// Marshal all to JSON.
	event, err := PrettyJSON(&streamEnvelope{
		Type: streamEnvelopeTypeEvent,
		Data: msg,
		Info: info,
	})
	if err != nil {
		return err
	}

	c.Logger().WithFields(logrus.Fields{
		"topics": topicDefinition.Topics,
		"id":     binder.id,
	}).Debugln("pubs: pub with connection")
	p.pubsub.Pub(event, topicDefinition.Topics...)
	return nil
}
