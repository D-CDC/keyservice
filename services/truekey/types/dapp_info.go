package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"ethereum/keyservice/accounts"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"ethereum/keyservice/crypto"
	"ethereum/keyservice/crypto/ecies"
	"ethereum/keyservice/log"
	"ethereum/keyservice/rlp"
	"io"
	"strconv"
	"time"
)

const (
	Unlock = 1
	Lock   = 0
)

type DappIdentify struct {
	ID           common.Hash
	Name         string
	IPs          []string
	Desc         string
	Priv         *ecdsa.PrivateKey
	Accounts     map[common.Address]*ChildAccount
	Session      *DappSession
	Status       uint64
	Create       common.Address
	Index        uint64
	AccountIndex uint64
}

type DappSession struct {
	CreatedAt uint64
	Key       []byte
	ID        common.Hash
}

func (ds *DappSession) Hash() common.Hash {
	return rlpHash([]interface{}{
		ds.CreatedAt,
		ds.Key,
	})
}

func NewDappIdentify(dapp DappQuest, root common.Address) (*DappIdentify, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, errors.New("generate key error")
	}
	di := &DappIdentify{
		Name:     dapp.Name,
		Desc:     dapp.Desc,
		Priv:     key,
		Status:   Unlock,
		Accounts: make(map[common.Address]*ChildAccount),
		Create:   root,
	}
	corectIp := CheckIp(dapp.IPs)
	di.IPs = make([]string, len(corectIp))
	copy(di.IPs, corectIp)

	return di, nil
}

func (di *DappIdentify) CheckSignature(hash, sign []byte) bool {
	pubKey, err := crypto.SigToPub(hash, sign)
	if err != nil {
		log.Error("crypto Message SigToPub error", "err", err)
		return false
	}
	address := crypto.PubkeyToAddress(*pubKey)

	if crypto.PubkeyToAddress(di.Priv.PublicKey) == address {
		return true
	}
	return false
}

func (di *DappIdentify) SignResult(pub *ecdsa.PublicKey, resultByte []byte) (*EncryptMessage, bool) {
	cryMessage := &EncryptMessage{
		CreatedAt: hexutil.Uint64(time.Now().Unix()),
	}
	encryptMessageInfo, err := ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(pub), resultByte, nil, nil)
	if err != nil {
		log.Error("publickey encrypt result error ", "publickey", common.Bytes2Hex(crypto.FromECDSAPub(pub)), "err", err)
		return nil, false
	}
	cryMessage.DappInfo = encryptMessageInfo
	hash := cryMessage.HashWithoutSign().Bytes()
	cryMessage.Sign, err = crypto.Sign(hash, di.Priv)
	return cryMessage, true
}

// "external" DappIdentify encoding. used for pos hd.
type extDappIdentify struct {
	ID           common.Hash
	Name         string
	IPs          []string
	Desc         string
	Priv         string
	Account      []*ChildAccount
	Status       uint64
	Create       common.Address
	Index        uint64
	AccountIndex uint64
}

func (i *DappIdentify) DecodeRLP(s *rlp.Stream) error {
	var ei extDappIdentify
	if err := s.Decode(&ei); err != nil {
		return err
	}
	aAccounts := make(map[common.Address]*ChildAccount)
	for _, account := range ei.Account {
		aAccounts[account.Account.Address] = account
	}
	privkey, err := crypto.HexToECDSA(ei.Priv)
	if err != nil {
		return err
	}
	i.ID, i.Name, i.IPs, i.Desc, i.Priv, i.Accounts, i.Status = ei.ID, ei.Name, ei.IPs, ei.Desc, privkey, aAccounts, ei.Status
	i.Create, i.Index, i.AccountIndex = ei.Create, ei.Index, ei.AccountIndex
	key := RandString(16)
	ds := &DappSession{
		CreatedAt: uint64(time.Now().Unix()),
		Key:       key,
	}
	ds.ID = ds.Hash()
	i.Session = ds
	return nil
}

// EncodeRLP serializes b into the truechain RLP AdminWallet format.
func (i *DappIdentify) EncodeRLP(w io.Writer) error {
	var aAccounts []*ChildAccount
	for _, k := range i.Accounts {
		aAccounts = append(aAccounts, k)
	}

	return rlp.Encode(w, extDappIdentify{
		ID:           i.ID,
		Name:         i.Name,
		IPs:          i.IPs,
		Desc:         i.Desc,
		Priv:         hex.EncodeToString(crypto.FromECDSA(i.Priv)),
		Account:      aAccounts,
		Status:       i.Status,
		Create:       i.Create,
		Index:        i.Index,
		AccountIndex: i.AccountIndex,
	})
}

type ChildAccount struct {
	ID         uint64           `json:"id"`
	Account    accounts.Account `json:"account"`
	PrivateKey *ecdsa.PrivateKey
}

func (c *ChildAccount) String() string {
	var ss string
	ss += "[ID:" + strconv.FormatUint(c.ID, 10)
	ss += " address:" + c.Account.Address.String() + " URL:" + c.Account.URL.String() + " ]"
	return ss
}

// "external" ChildAccount encoding. used for pos hd.
type extChildAccount struct {
	ID      uint64           `json:"id"`
	Account accounts.Account `json:"account"`
}

func (i *ChildAccount) DecodeRLP(s *rlp.Stream) error {
	var ei extChildAccount
	if err := s.Decode(&ei); err != nil {
		return err
	}
	i.ID, i.Account = ei.ID, ei.Account
	return nil
}

// EncodeRLP serializes b into the truechain RLP AdminWallet format.
func (i *ChildAccount) EncodeRLP(w io.Writer) error {

	return rlp.Encode(w, extChildAccount{
		ID:      i.ID,
		Account: i.Account,
	})
}
