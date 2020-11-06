// Copyright 2019 The go-ethereum Authors
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
//

package signer

import (
	"context"
	"encoding/json"
	"ethereum/keyservice/accounts"
	"ethereum/keyservice/accounts/keystore"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"ethereum/keyservice/rlp"
	"ethereum/keyservice/services/truekey/hdwallet"
	"ethereum/keyservice/services/truekey/types"
	"fmt"
	"golang.org/x/crypto/sha3"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SignerUIAPI implements methods truekey provides for a UI to query, in the bidirectional communication
// channel.
// This API is considered secure, since a request can only
// ever arrive from the UI -- and the UI is capable of approving any action, thus we can consider these
// requests pre-approved.
// NB: It's very important that these methods are not ever exposed on the external service
// registry.
type UIServerAPI struct {
	extApi *SignerAPI
}

// NewUIServerAPI creates a new UIServerAPI
func NewUIServerAPI(extapi *SignerAPI) *UIServerAPI {
	return &UIServerAPI{extapi}
}

// DeriveAccounts requests a HD wallet to derive a new account, optionally pinning
// it for later reuse.
// Example call
// {"jsonrpc":"2.0","method":"truekey_registerDapp","params":["ledger://","m/44'/60'/0'", false], "id":6}
func (s *UIServerAPI) RegisterDapp(ctx context.Context, quest types.AdminQuest, encryMessage types.EncryptMessage) (*types.EncryptMessage, error) {
	//return s.extApi.registerDapp(quest, encryMessage, MetadataFromContext(ctx))
	return nil, nil
}

func (s *UIServerAPI) RegisterAccount(ctx context.Context, phone string) (common.Address, error) {
	var phoneNumber types.Phone
	err := json.Unmarshal([]byte(phone), &phoneNumber)
	if err != nil {
		return common.Address{}, err
	}

	return s.extApi.register(uint64(phoneNumber.Phone))
}

// List available accounts. As opposed to the external API definition, this method delivers
// the full Accounts object and not only Address.
// Example call
// {"jsonrpc":"2.0","method":"truekey_listAccounts","params":[], "id":4}
func (s *UIServerAPI) AuthPub(ctx context.Context, quest types.AdminQuest, auth types.AuthQuest) (*types.EncryptMessage, error) {
	//return s.extApi.authPub(quest, auth, MetadataFromContext(ctx))
	return nil, nil
}

func (s *UIServerAPI) SignHash(ctx context.Context, key common.Hash, addr common.Address, id common.Hash, encryMessage types.EncryptMessage) (*types.EncryptMessage, error) {
	//return s.extApi.SignHash(ctx,key,addr,id,encryMessage)
	return nil, nil
}

func (s *UIServerAPI) SignHashPlain(ctx context.Context, txStr string) (hexutil.Bytes, error) {
	var tx types.SignTx
	err := json.Unmarshal([]byte(txStr), &tx)
	if err != nil {
		return nil, err
	}
	return s.extApi.SignHashPlain(ctx, tx.Phone, tx)
}

func (s *UIServerAPI) Version(ctx context.Context) (string, error) {
	return s.extApi.Version(ctx)
}

func getKeyStoreDir(root string, hash common.Hash) string {
	return filepath.Join(root, common.Bytes2Hex(hash.Bytes()))
}

//"m/44'/60'/0'/0/"
func GetDerivationPath(phone uint64) (accounts.DerivationPath, error) {
	var pathIndex string
	if phone > 4294967290 {
		phoneStr := strconv.FormatUint(phone, 10)
		dappIndex, err := strconv.ParseUint(phoneStr[:6], 10, 64)
		if err != nil {
			return nil, types.ErrPhoneNumberError
		}
		index, err := strconv.ParseUint(phoneStr[6:], 10, 64)
		if err != nil {
			return nil, types.ErrPhoneNumberError
		}
		arrs := strings.Split(DefaultBaseDerivationPath, "/")
		arrs[3] = strconv.FormatInt(int64(dappIndex), 10) + "'"
		subDerivationPath := strings.Join(arrs, "/")
		pathIndex = subDerivationPath + fmt.Sprintf("%d", index)
	} else {
		arrs := strings.Split(DefaultBaseDerivationPath, "/")
		arrs[3] = strconv.FormatInt(int64(0), 10) + "'"
		subDerivationPath := strings.Join(arrs, "/")
		pathIndex = subDerivationPath + fmt.Sprintf("%d", phone)
	}
	fmt.Println(pathIndex)
	return hdwallet.MustParseDerivationPath(pathIndex), nil
}

func verifyPhone(phone uint64) bool {
	reg := `^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`
	rgx := regexp.MustCompile(reg)
	phoneNumber := strconv.FormatInt(int64(phone), 10) + "'"

	return rgx.MatchString(phoneNumber)
}

func startTrueKeyKeyStore(ksLocation string, lightKDF bool) *keystore.KeyStore {
	var (
		n, p = keystore.StandardScryptN, keystore.StandardScryptP
	)
	if lightKDF {
		n, p = keystore.LightScryptN, keystore.LightScryptP
	}
	// TrueKey doesn't allow insecure http account unlock.
	return keystore.NewKeyStore(ksLocation, n, p)
}

// Metadata about a request
type Metadata struct {
	Remote    string `json:"remote"`
	Local     string `json:"local"`
	Scheme    string `json:"scheme"`
	UserAgent string `json:"User-Agent"`
	Origin    string `json:"Origin"`
}

// MetadataFromContext extracts Metadata from a given context.Context
func MetadataFromContext(ctx context.Context) Metadata {
	m := Metadata{"NA", "NA", "NA", "", ""} // batman

	if v := ctx.Value("remote"); v != nil {
		m.Remote = v.(string)
	}
	if v := ctx.Value("scheme"); v != nil {
		m.Scheme = v.(string)
	}
	if v := ctx.Value("local"); v != nil {
		m.Local = v.(string)
	}
	if v := ctx.Value("Origin"); v != nil {
		m.Origin = v.(string)
	}
	if v := ctx.Value("User-Agent"); v != nil {
		m.UserAgent = v.(string)
	}
	return m
}

// String implements Stringer interface
func (m Metadata) String() string {
	s, err := json.Marshal(m)
	if err == nil {
		return string(s)
	}
	return err.Error()
}

func (m Metadata) Hash() common.Hash {
	return rlpHash(m)
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
