// Copyright 2018 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package signer

import (
	"context"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"ethereum/keyservice/log"
	"ethereum/keyservice/services/truekey/types"
)

type ServerAuditLogger struct {
	log log.Logger
	api types.ServerAPI
}

func (l *ServerAuditLogger) RegisterDapp(ctx context.Context, quest types.AdminQuest, encryMessage types.EncryptMessage) (*types.EncryptMessage, error) {
	l.log.Info("RegisterDapp", "type", "request", "metadata", MetadataFromContext(ctx).String(), "quest", quest, "encryMessage", encryMessage)
	res, e := l.api.RegisterDapp(ctx, quest, encryMessage)
	l.log.Info("RegisterDapp", "type", "response", "data", res, "error", e)
	return res, e
}

func (l *ServerAuditLogger) RegisterAccount(ctx context.Context, phone string) (common.Address, error) {
	l.log.Info("RegisterAccount", "type", "request", "metadata", MetadataFromContext(ctx).String(), "quest", phone)
	res, e := l.api.RegisterAccount(ctx, phone)
	l.log.Info("RegisterAccount", "type", "response", "data", res, "error", e)
	return res, e
}

func (l *ServerAuditLogger) AuthPub(ctx context.Context, quest types.AdminQuest, auth types.AuthQuest) (*types.EncryptMessage, error) {
	l.log.Info("AuthPub", "type", "request", "metadata", MetadataFromContext(ctx).String(),
		"quest", quest,
		"auth", auth,
	)
	res, err := l.api.AuthPub(ctx, quest, auth)
	l.log.Info("AuthPub", "type", "response", "res", res, "error", err)
	return res, err
}

func (l *ServerAuditLogger) SignHash(ctx context.Context, key common.Hash, addr common.Address, id common.Hash, encryMessage types.EncryptMessage) (*types.EncryptMessage, error) {
	l.log.Info("SignHash", "type", "request", "metadata", MetadataFromContext(ctx).String(),
		"key", key.String(),
		"addr", addr.String(),
		"id", id.String(),
		"encryMessage", encryMessage)

	res, e := l.api.SignHash(ctx, key, addr, id, encryMessage)
	l.log.Info("SignHash", "type", "response", "data", res, "error", e)
	return res, e
}

func (l *ServerAuditLogger) SignHashPlain(ctx context.Context, query string) (hexutil.Bytes, error) {
	l.log.Info("SignHashPlain", "type", "request", "metadata", MetadataFromContext(ctx).String(),
		"encryMessage", query)

	res, e := l.api.SignHashPlain(ctx, query)
	l.log.Info("SignHashPlain", "type", "response", "data", res, "error", e)
	return res, e
}

func (l *ServerAuditLogger) Version(ctx context.Context) (string, error) {
	l.log.Info("Version", "type", "request", "metadata", MetadataFromContext(ctx).String())
	data, err := l.api.Version(ctx)
	l.log.Info("Version", "type", "response", "data", data, "error", err)
	return data, err

}

func NewServerAuditLogger(path string, api types.ServerAPI) (*ServerAuditLogger, error) {
	l := log.New("api", "signer")
	handler, err := log.FileHandler(path, log.LogfmtFormat())
	if err != nil {
		return nil, err
	}
	l.SetHandler(handler)
	l.Info("Configured", "server audit log", path)
	return &ServerAuditLogger{l, api}, nil
}
