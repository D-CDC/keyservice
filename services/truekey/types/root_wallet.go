package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"ethereum/keyservice/crypto"
	"ethereum/keyservice/crypto/ecies"
	"ethereum/keyservice/log"
	"ethereum/keyservice/rlp"
	"ethereum/keyservice/services/truekey/hdwallet"
	"fmt"
	"strings"
	"time"
)

type RootWallet struct {
	Wallet   *hdwallet.Wallet
	Accounts map[uint64]*ChildAccount
}

func addressEqual(address1, address2 common.Address) bool {
	if strings.ToLower(address1.String()) == strings.ToLower(address2.String()) {
		return true
	}
	return false
}

//	pubKey, err := crypto.SigToPub(hashBytes, encryptNode.Sign)
type AdminWallet struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
	DappPub    *ecdsa.PublicKey
}

func NewAdminWallet(address common.Address) (*AdminWallet, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return &AdminWallet{
		Address:    address,
		PrivateKey: key,
	}, nil
}

func (aw *AdminWallet) CheckSignature(hash, sign []byte) bool {
	pubKey, err := crypto.SigToPub(hash, sign)
	if err != nil {
		log.Error("crypto Message SigToPub error", "err", err)
		return false
	}
	address := crypto.PubkeyToAddress(*pubKey)
	fmt.Println("CheckSignature address", address.String())
	if addressEqual(address, aw.Address) {
		aw.DappPub = pubKey
		return true
	}
	return false
}

func (aw *AdminWallet) SignResult(val interface{}) (*EncryptMessage, error) {
	resultByte, err := rlp.EncodeToBytes(val)
	if err != nil {
		log.Error("EncodeToBytes error: ", "err", err)
		return nil, err
	}
	cryMessage, ok := aw.signResult(resultByte)

	if !ok {
		return nil, ErrAdminSignError
	}
	return cryMessage, nil
}

func (aw *AdminWallet) signResult(resultByte []byte) (*EncryptMessage, bool) {
	cryMessage := &EncryptMessage{
		CreatedAt: hexutil.Uint64(time.Now().Unix()),
	}
	encryptMessageInfo, err := ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(aw.DappPub), resultByte, nil, nil)
	if err != nil {
		log.Error("publickey encrypt result error ", "publickey", common.Bytes2Hex(crypto.FromECDSAPub(aw.DappPub)), "err", err)
		return nil, false
	}
	cryMessage.DappInfo = encryptMessageInfo
	hash := cryMessage.HashWithoutSign().Bytes()
	cryMessage.Sign, err = crypto.Sign(hash, aw.PrivateKey)
	return cryMessage, true
}

// which encrypt message with member Publickey
//EncryptMessage  all information of the dapp
type EncryptMessage struct {
	CreatedAt hexutil.Uint64 `json:"create_at"`
	DappInfo  hexutil.Bytes  `json:"dapp_info"`
	Sign      hexutil.Bytes  `json:"sign"`
}

// which encrypt message with member Publickey
//EncryptMessage  all information of the dapp
type EncryptAuth struct {
	CreatedAt hexutil.Uint64 `json:"create_at"`
	DappInfo  hexutil.Bytes  `json:"dapp_info"`
	Sign      hexutil.Bytes  `json:"sign"`
	ID        common.Hash    `json:"id"`
}

func (c *EncryptMessage) String(str string) {
	log.Info(str, "reatedAt", c.CreatedAt, "dappinfo", common.Bytes2Hex(c.DappInfo), "sign", common.Bytes2Hex(c.Sign))
}

func (c *EncryptMessage) HashWithoutSign() common.Hash {
	return rlpHash([]interface{}{
		c.CreatedAt,
		c.DappInfo,
	})
}

func (c *EncryptAuth) String(str string) {
	log.Info(str, "reatedAt", c.CreatedAt, "dappinfo", common.Bytes2Hex(c.DappInfo), "sign", common.Bytes2Hex(c.Sign), "id", c.ID.String())
}

func (c *EncryptAuth) HashWithoutSign() common.Hash {
	return rlpHash([]interface{}{
		c.CreatedAt,
		c.DappInfo,
	})
}
