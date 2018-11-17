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
	"io/ioutil"
	"net/http"
	"strings"

	"stash.kopano.io/kc/kapi/auth"
	"stash.kopano.io/kc/kapi/plugins/kvs/kv"
)

const valueSizeLimit = 16384

func (p *KVSPlugin) handleUserKV(rw http.ResponseWriter, req *http.Request) {
	key := req.URL.Path

	user, ok := auth.RecordFromContext(req.Context())
	if !ok {
		http.Error(rw, "", http.StatusForbidden)
		return
	}

	req.ParseForm()

	switch req.Method {
	case http.MethodGet:
		p.handleGet(rw, req, "user", key, user)
		return
	case http.MethodPut:
		p.handleCreateOrUpdate(rw, req, "user", key, user)
		return
	case http.MethodDelete:
		p.handleDelete(rw, req, "user", key, user)
		return
	}

	http.Error(rw, "", http.StatusNotImplemented)
}

func (p *KVSPlugin) handleGet(rw http.ResponseWriter, req *http.Request, realm string, key string, user *auth.Record) {
	recurse := req.Form.Get("recurse") == "1"
	raw := req.Form.Get("raw") == "1"

	var collection *string
	parts := strings.SplitN(key, "/", 2)
	if len(parts) > 0 && recurse {
		collection = &parts[0]
	}

	record := &kv.Record{
		Collection: collection,
		Key:        key,
		OwnerID:    user.AuthenticatedUserID,
		ClientID:   user.StandardClaims.Audience,
	}

	result, err := p.store.Get(req.Context(), realm, record)
	if err != nil {
		p.srv.Logger().Debugf("kvs: failed to get from kv: %v", err)
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}

	if len(result) == 0 {
		if !recurse {
			http.NotFound(rw, req)
			return
		}
		// Fake collection.
		result = append(result, &kv.Record{
			Key:         key,
			ContentType: "application/json",
			Value:       []byte("[]"),
		})
	}

	if len(result) > 1 {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("[\n"))
		first := true
		filter := len(parts) > 1 // Filter further.
		for _, r := range result {
			if filter && !strings.HasPrefix(r.Key, key) {
				continue
			}
			d, err := r.EncodeToJSON()
			if err != nil {
				p.srv.Logger().WithField("key", r.Key).Warnf("kvs: failed to JSON encode record: %v", err)
				continue
			}
			if first {
				first = false
			} else {
				rw.Write([]byte(",\n"))
			}
			rw.Write(d)
		}
		rw.Write([]byte("]\n"))
	} else {
		r := result[0]
		var d []byte
		if raw {
			if r.ContentType != "" {
				rw.Header().Set("Content-Type", r.ContentType)
			} else {
				rw.Header().Set("Content-Type", "application/octet-stream")
			}
			d = r.Value
		} else {
			var err error
			d, err = r.EncodeToJSON()
			if err != nil {
				p.srv.Logger().WithField("key", r.Key).Warnf("kvs: failed to JSON encode record: %v", err)
				http.Error(rw, "", http.StatusInternalServerError)
				return
			}
			rw.Header().Set("Content-Type", "application/json")
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(d)
	}

}

func (p *KVSPlugin) handleCreateOrUpdate(rw http.ResponseWriter, req *http.Request, realm string, key string, user *auth.Record) {
	req.Body = http.MaxBytesReader(rw, req.Body, valueSizeLimit)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		p.srv.Logger().Debugf("kvs: failed to read request body: %v", err)
		http.Error(rw, "", http.StatusBadRequest)
		return
	}

	contentType := req.Header.Get("Content-Type")
	var collection string
	parts := strings.SplitN(key, "/", 2)
	if len(parts) > 1 {
		collection = parts[0]
	}

	record := &kv.Record{
		Collection:  &collection,
		Key:         key,
		Value:       body,
		ContentType: contentType,
		OwnerID:     user.AuthenticatedUserID,
		ClientID:    user.StandardClaims.Audience,
	}

	err = p.store.CreateOrUpdate(req.Context(), realm, record)
	if err != nil {
		p.srv.Logger().Debugf("kvs: failed to create or update from kv: %v", err)
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (p *KVSPlugin) handleDelete(rw http.ResponseWriter, req *http.Request, realm string, key string, user *auth.Record) {
	record := &kv.Record{
		Key:      key,
		OwnerID:  user.AuthenticatedUserID,
		ClientID: user.StandardClaims.Audience,
	}

	ok, err := p.store.Delete(req.Context(), realm, record)
	if err != nil {
		p.srv.Logger().Debugf("kvs: failed to delete from kv: %v", err)
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}

	if ok {
		rw.WriteHeader(http.StatusOK)
	} else {
		rw.WriteHeader(http.StatusNotFound)
	}
}
