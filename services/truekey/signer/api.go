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
	"crypto/ecdsa"
	"ethereum/keyservice/accounts/keystore"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	coreType "ethereum/keyservice/core/types"
	"ethereum/keyservice/crypto"
	"ethereum/keyservice/etruedb"
	"ethereum/keyservice/log"
	"ethereum/keyservice/rlp"
	"ethereum/keyservice/services/truekey/hdwallet"
	"ethereum/keyservice/services/truekey/rawdb"
	"ethereum/keyservice/services/truekey/types"
	"fmt"
	"math/big"
	"os"
	"sync"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

const (
	DefaultBaseDerivationPath = "m/44'/60'/0'/0/"
)

// SignerAPI defines the actual implementation of ExternalAPI
type SignerAPI struct {
	db          etruedb.Database
	rootWallets map[common.Address]*types.RootWallet
	indexMutex  *sync.Mutex //block mutex
	PrivateKeys map[common.Address]*ecdsa.PrivateKey
}

// NewSignerAPI creates a new API that can be used for Accounts management.
// ksLocation specifies the directory where to store the password protected private
// key that is generated when a new Accounts is created.
func NewSignerAPI(db etruedb.Database, keys []*keystore.Key, configs []types.RootConfig) (*SignerAPI, error) {

	signer := &SignerAPI{
		db:          db,
		rootWallets: make(map[common.Address]*types.RootWallet),
		indexMutex:  new(sync.Mutex),
		PrivateKeys: make(map[common.Address]*ecdsa.PrivateKey),
	}
	for _, k := range keys {
		wallet, err := hdwallet.NewFromSeed(crypto.FromECDSA(k.PrivateKey))
		if err != nil {
			return nil, err
		}
		address := crypto.PubkeyToAddress(k.PrivateKey.PublicKey)
		log.Info("NewSignerAPI", "address", address)
		signer.PrivateKeys[address] = k.PrivateKey
		signer.rootWallets[k.Address] = &types.RootWallet{
			Wallet:   wallet,
			Accounts: make(map[uint64]*types.ChildAccount),
		}
	}
	signer.init(configs)
	return signer, nil
}

func (api *SignerAPI) init(configs []types.RootConfig) {
	if len(configs) > 0 {
		for _, root := range configs {
			v, exists := api.rootWallets[root.Root]
			if !exists {
				continue
			}
			childHashs := rawdb.ReadRootInfo(api.db, root.Root.Hash())
			for _, hash := range childHashs {
				child := rawdb.ReadChildAccount(api.db, hash)
				privateKey, err := v.Wallet.PrivateKey(child.Account)
				if err != nil {
					fmt.Println(fmt.Sprintf("%v: %v", "Wallet calculate PrivateKey error", err))
					continue
				}
				child.PrivateKey = privateKey
				v.Accounts[convertHashToUint(hash)] = child
			}
		}
	}
}

func convertBigToHash(uint642 uint64) common.Hash {
	return common.BigToHash(new(big.Int).SetUint64(uint642))
}

func convertHashToUint(hash common.Hash) uint64 {
	return new(big.Int).SetBytes(hash.Bytes()).Uint64()
}

//func (api *SignerAPI) registerDapp(quest types.AdminQuest, encryMessage types.EncryptMessage, metadata Metadata) (*types.EncryptMessage, error) {
//	api.indexMutex.Lock()
//	defer api.indexMutex.Unlock()
//	rootWallet, err := api.checkAdmin(quest)
//	if err != nil {
//		return nil, err
//	}
//
//	priKey := ecies.ImportECDSA(adminWallet.PrivateKey)
//	decryptMessage, err := priKey.Decrypt(encryMessage.DappInfo, nil, nil)
//	dappquest := new(types.DappQuest)
//
//	if err := rlp.DecodeBytes(decryptMessage, dappquest); err != nil {
//		fmt.Println("Failed to decode decrypt message", "err", err)
//		return nil, err
//	}
//
//	if err != nil {
//		return nil, err
//	}
//	return nil,nil
//}

func (api *SignerAPI) register(phone uint64) (common.Address, error) {
	api.indexMutex.Lock()
	defer api.indexMutex.Unlock()
	root := common.HexToAddress("0xe4FAd2E5eE2E878e65F1fe02c0F9edAf54789a8e")
	v, err := api.checkRoot(root)
	if err != nil {
		return common.Address{}, err
	}
	child, _ := api.checkChildExist(phone, root)
	if child != nil {
		return child.Account.Address, nil
	}

	childAccount, err := api.getChild(phone, v)
	if err != nil {
		return common.Address{}, nil
	}

	rawdb.WriteChildAccount(api.db, convertBigToHash(phone), v.Accounts[phone])
	var ids []common.Hash
	if rawdb.HasRootInfo(api.db, root.Hash()) {
		ids = append(ids, rawdb.ReadRootInfo(api.db, root.Hash())...)
	}
	rawdb.WriteRootInfo(api.db, root.Hash(), ids)
	log.Info("register", "phone", phone, "address", childAccount.Account.Address.String())
	return childAccount.Account.Address, nil
}

func (api *SignerAPI) getChild(phone uint64, v *types.RootWallet) (*types.ChildAccount, error) {
	path, err := GetDerivationPath(phone)
	if err != nil {
		return nil, err
	}

	accountHD, err := v.Wallet.Derive(path, false)
	if err != nil {
		log.Info("Derive accounts", "err", err)
		return nil, err
	}
	privateKey, err := v.Wallet.PrivateKey(accountHD)
	if err != nil {
		log.Info("Derive accounts PrivateKey", "err", err)
		return nil, err
	}
	v.Accounts[phone] = &types.ChildAccount{
		ID:         phone,
		Account:    accountHD,
		PrivateKey: privateKey,
	}
	return v.Accounts[phone], nil
}

func (api *SignerAPI) checkAdmin(quest types.AdminQuest) (*types.RootWallet, error) {
	v, exists := api.rootWallets[quest.Root]
	if !exists {
		return nil, types.ErrRootError
	}

	return v, nil
}

func (api *SignerAPI) checkRoot(root common.Address) (*types.RootWallet, error) {
	v, exists := api.rootWallets[root]
	if !exists {
		return nil, types.ErrRootError
	}

	return v, nil
}

func (api *SignerAPI) checkChildExist(id uint64, root common.Address) (*types.ChildAccount, error) {
	dapp, find := api.rootWallets[root].Accounts[id]
	if find {
		return dapp, types.ErrDappNotRegister
	}
	return nil, types.ErrDappNotRegister
}

//func (api *SignerAPI) authPub(quest types.AdminQuest, auth types.AuthQuest, metadata Metadata) (*types.EncryptMessage, error) {
//	api.indexMutex.Lock()
//	defer api.indexMutex.Unlock()
//
//	rootWallet, err := api.checkAdmin(quest)
//	if err != nil {
//		return nil, err
//	}
//
//	if !adminWallet.CheckSignature(auth.Hash.Bytes(), auth.Sign) {
//		return nil, types.ErrSignTxError
//	}
//	ar := &types.AuthResult{
//		CryptoPub: hexutil.Encode(crypto.FromECDSAPub(&adminWallet.PrivateKey.PublicKey)),
//	}
//
//	cryMessage, err := adminWallet.SignResult(ar)
//	if err != nil {
//		return nil, err
//	}
//	return cryMessage, nil
//}

// -------------------------------------------------------------------------------

//func (api *SignerAPI) SignHash(ctx context.Context, key common.Hash, addr common.Address, id common.Hash, encryMessage types.EncryptMessage) (*types.EncryptMessage, error) {
//	api.indexMutex.Lock()
//	defer api.indexMutex.Unlock()
//
//	if timestamp := time.Unix(int64(encryMessage.CreatedAt), 0); time.Since(timestamp) > time.Minute*30 {
//		return nil, errors.New(fmt.Sprintf("Sign timeout 30 minute  %v", timestamp))
//	}
//
//	dapp, exists := api.dapps[key]
//	if !exists {
//		return nil, types.ErrChildNotExist
//	}
//	_, exist := api.rootWallets[dapp.Create]
//	if !exist {
//		return nil, types.ErrRootNotServer
//	}
//
//	if timestamp := time.Unix(int64(dapp.Session.CreatedAt), 0); time.Since(timestamp) > time.Hour*24*30 {
//		return nil, errors.New(fmt.Sprintf("Session timeout 30 day  %v please auth dapp again", timestamp))
//	}
//
//	if dapp.Session.ID != id {
//		return nil, errors.New(fmt.Sprintf("Session id error  %v, please auth dapp again", id))
//	}
//
//	data, err := crypto.AESCbCDecrypt(encryMessage.Sign, dapp.Session.Key)
//	if err != nil {
//		return nil, errors.New(fmt.Sprintf("Session id error  %v", err))
//	}
//
//	account, exists := dapp.Accounts[addr]
//	if !exists {
//		return nil, types.ErrAccountNotExist
//	}
//	metadata := MetadataFromContext(ctx)
//
//	if account.Status == types.Lock {
//		return nil, types.ErrAccountLock
//	}
//	find := false
//	for _, aip := range account.IPs {
//		if types.ContainIp(metadata.Remote, aip) {
//			find = true
//		}
//	}
//	if !find {
//		return nil, types.ErrAccountSign
//	}
//
//	data, err = hex.DecodeString(string(data))
//	if err != nil {
//		return nil, errors.New(fmt.Sprintf("%v: %v", "Sign decode string error", err))
//	}
//
//	result, err := crypto.Sign(data, account.PrivateKey)
//	if err != nil {
//		return nil, errors.New(fmt.Sprintf("%v: %v", "Sign error", err))
//	}
//
//	cryMessage := &types.EncryptMessage{
//		CreatedAt: hexutil.Uint64(time.Now().Unix()),
//	}
//	//cryMessage.DappInfo = result
//	//cryMessage.Sign = crypto.Keccak256Hash(crypto.AESCbCEncrypt([]byte(dapp.Session.Key), result)).Bytes()
//	cryMessage.Sign, err = crypto.AESCbCEncrypt([]byte(hex.EncodeToString(result)), dapp.Session.Key)
//	if err != nil {
//		return nil, errors.New(fmt.Sprintf("AESCbCEncrypt session id error  %v", err))
//	}
//	return cryMessage, nil
//}

func (api *SignerAPI) SignHashPlain(ctx context.Context, phone uint64, tx types.SignTx) (hexutil.Bytes, error) {
	api.indexMutex.Lock()
	defer api.indexMutex.Unlock()
	root := common.HexToAddress("0xe4FAd2E5eE2E878e65F1fe02c0F9edAf54789a8e")
	dapp, exists := api.rootWallets[root]
	if !exists {
		return nil, types.ErrRootNotServer
	}

	account, exists := dapp.Accounts[phone]
	if !exists {
		var err error
		account, err = api.getChild(phone, dapp)
		if err != nil {
			return nil, types.ErrAccountNotExist
		}
	}
	var transaction *coreType.Transaction
	var err error
	sender := coreType.NewTIP1Signer(new(big.Int).SetUint64(tx.ChainId))
	if tx.Payment != (common.Address{}) {
		v, ok := api.PrivateKeys[tx.Payment]
		if !ok {
			return nil, types.ErrPaymentError
		}
		transaction = coreType.NewTransaction_Payment(tx.Nonce, tx.To, tx.Value, new(big.Int).SetUint64(0), tx.GasLimit, new(big.Int).SetUint64(tx.GasPrice), tx.Data, tx.Payment)
		transaction, err = coreType.SignTx(transaction, sender, account.PrivateKey)
		transaction, err = coreType.SignTx_Payment(transaction, sender, v)
		if err != nil {
			return nil, types.ErrSignTxError
		}
	} else {
		transaction = coreType.NewTransaction(tx.Nonce, tx.To, tx.Value, tx.GasLimit, new(big.Int).SetUint64(tx.GasPrice), tx.Data)
		transaction, err = coreType.SignTx(transaction, sender, account.PrivateKey)
		if err != nil {
			return nil, types.ErrSignTxError
		}
	}
	if transaction == nil {
		return nil, types.ErrCreateTxError
	}
	data, err := rlp.EncodeToBytes(transaction)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Returns the external api version. This method does not require user acceptance. Available methods are
// available via enumeration anyway, and this info does not contain user-specific data
func (api *SignerAPI) Version(ctx context.Context) (string, error) {
	return types.ExternalAPIVersion, nil
}

func (api *SignerAPI) Stop() {
	var rootInfo []common.Hash
	for _, v := range api.rootWallets {
		for k, account := range v.Accounts {
			if !rawdb.HasChildAccount(api.db, convertBigToHash(k)) {
				rawdb.WriteChildAccount(api.db, convertBigToHash(k), account)
			}
			rootInfo = append(rootInfo, convertBigToHash(k))
		}
	}
	for k, _ := range api.rootWallets {
		rawdb.WriteRootInfo(api.db, k.Hash(), rootInfo)
	}

	log.Info("Signer stop")
}
