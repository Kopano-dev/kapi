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

package main

import (
	"net/http"
	"strings"

	rpcpb "github.com/cs3org/go-cs3apis/cs3/rpc"
	storageproviderv0alphapb "github.com/cs3org/go-cs3apis/cs3/storageprovider/v0alpha"
	"github.com/cs3org/reva/pkg/token"
	"google.golang.org/grpc/metadata"
)

const defaultHeader = "x-access-token"

func GetToken(r *http.Request) string {
	// 1. check Authorization header
	hdr := r.Header.Get("Authorization")
	token := strings.TrimPrefix(hdr, "Bearer ")
	if token != "" {
		return token
	}
	// TODO 2. check form encoded body parameter for POST requests, see https://tools.ietf.org/html/rfc6750#section-2.2

	// 3. check uri query parameter, see https://tools.ietf.org/html/rfc6750#section-2.3
	tokens, ok := r.URL.Query()["access_token"]
	if !ok || len(tokens[0]) < 1 {
		return ""
	}

	return tokens[0]
}

func (p *OcisPlugin) listFiles(rw http.ResponseWriter, request *http.Request) {
	accessToken := GetToken(request)
	ctx := request.Context()
	ctx = token.ContextSetToken(ctx, accessToken)
	ctx = metadata.AppendToOutgoingContext(ctx, defaultHeader, accessToken)
	p.srv.Logger().Warnf("owncloud-plugin: provides access token %s", ctx)

	// TODO: read the path from request
	fn := "/"
	listChildren := true

	// TODO: where to get the client from
	client, err := p.getClient()
	if err != nil {
		p.srv.Logger().WithError(err).Warn("owncloud-plugin: error getting grpc client")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	ref := &storageproviderv0alphapb.Reference{
		Spec: &storageproviderv0alphapb.Reference_Path{Path: fn},
	}
	req := &storageproviderv0alphapb.StatRequest{Ref: ref}
	res, err := client.Stat(ctx, req)
	if err != nil {
		p.srv.Logger().WithError(err).Warn("owncloud-plugin: error sending a grpc stat request")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if res.Status.Code != rpcpb.Code_CODE_OK {
		if res.Status.Code == rpcpb.Code_CODE_NOT_FOUND {
			p.srv.Logger().WithError(err).Warnf("owncloud-plugin: resource not found %s", fn)
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	info := res.Info
	infos := []*storageproviderv0alphapb.ResourceInfo{info}
	if info.Type == storageproviderv0alphapb.ResourceType_RESOURCE_TYPE_CONTAINER && listChildren {
		req := &storageproviderv0alphapb.ListContainerRequest{
			Ref: ref,
		}
		res, err := client.ListContainer(ctx, req)
		if err != nil {
			p.srv.Logger().WithError(err).Warnf("owncloud-plugin: error sending list container grpc request %s", fn)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if res.Status.Code != rpcpb.Code_CODE_OK {
			p.srv.Logger().WithError(err).Warnf("owncloud-plugin: error calling grpc list container %s", fn)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		infos = append(infos, res.Infos...)
	}

	js, err := p.formatDriveItems(infos)
	if err != nil {
		p.srv.Logger().Errorf("owncloud-plugin: error encoding response as json %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(js)
}
