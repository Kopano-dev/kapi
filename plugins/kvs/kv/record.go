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

package kv

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
)

// A Record holds the data stored in a KV.
type Record struct {
	Collection  *string
	Key         string
	Value       []byte
	ContentType string

	OwnerID  string
	ClientID string

	RequiredScopes string
}

// A RecordJSON is the public JSON representation fo a Record.
type RecordJSON struct {
	Key         *string         `json:"key"`
	Value       json.RawMessage `json:"value"`
	ContentType string          `json:"content_type,omitempty"`
}

// EncodeToJSON encodes the accociated Record to JSON.
func (r *Record) EncodeToJSON() (json.RawMessage, error) {
	j := &RecordJSON{
		Key: &r.Key,
	}

	if strings.HasPrefix(r.ContentType, "application/json") {
		j.Value = json.RawMessage(r.Value)
	} else {
		buf := bytes.NewBufferString("\"")
		encoder := base64.NewEncoder(base64.StdEncoding, buf)
		_, err := encoder.Write(r.Value)
		if err != nil {
			return nil, err
		}
		encoder.Close()
		buf.WriteString("\"")
		j.Value = buf.Bytes()
		j.ContentType = r.ContentType
	}

	return json.MarshalIndent(j, "", "  ")
}
