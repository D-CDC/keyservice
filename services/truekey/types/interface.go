package types

import (
	"context"
	"encoding/json"
	"errors"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const (
	// ExternalAPIVersion
	ExternalAPIVersion = "1.0.0"
)

type ClientQuest struct {
	ID        common.Address `json:"id"`
	AddressId common.Address `json:"address_id"`
}

type ClientQuestParam struct {
	ID        common.Address `json:"timestamp"`
	AddressId common.Address `json:"address_id"`
}

type ClientResult struct {
	Status     string `json:"status"`
	SignResult []byte `json:"data"`
}

type ClentQuest struct {
	Data string `json:"data"`
	Pub  string `json:"pub"`
}

// ServerAPI defines the admin API through which control hd  wallet.
type ServerAPI interface {
	// Register a admin
	RegisterDapp(ctx context.Context, quest AdminQuest, encryMessage EncryptMessage) (*EncryptMessage, error)
	// Register a account
	RegisterAccount(ctx context.Context, phone string) (common.Address, error)
	// auth admin
	AuthPub(ctx context.Context, quest AdminQuest, auth AuthQuest) (*EncryptMessage, error)
	// SignHash request to sign the specified transaction
	SignHash(ctx context.Context, dappid common.Hash, addr common.Address, id common.Hash, encryMessage EncryptMessage) (*EncryptMessage, error)
	// SignHash request to sign the specified hash no crypto data , data hexutil.Bytes ClentQuest
	SignHashPlain(ctx context.Context, tx string) (hexutil.Bytes, error)
	// Version info about the APIs
	Version(ctx context.Context) (string, error)
}

type Phone struct {
	Phone int64 `json:"userId"`
}

type SignTx struct {
	Phone    uint64         `json:"userId"`
	To       common.Address `json:"to"`
	Value    *big.Int       `json:"value"`
	GasPrice uint64         `json:"gasPrice"`
	GasLimit uint64         `json:"gasLimit"`
	Nonce    uint64         `json:"nonce"`
	Data     []byte         `json:"data"`
	ChainId  uint64         `json:"chainId"`
	Payment  common.Address `json:"payment"`
}

//// MarshalJSON marshals as JSON.
//func (t SignTx) MarshalJSON() ([]byte, error) {
//	type SignTx struct {
//		Phone    hexutil.Uint64  `json:"phone"`
//		To       *common.Address `json:"to"`
//		Value    *hexutil.Big    `json:"value"`
//		GasPrice hexutil.Uint64  `json:"gas_price"`
//		GasLimit hexutil.Uint64  `json:"gas_limit"`
//		Nonce    hexutil.Uint64  `json:"nonce"`
//		Data     hexutil.Bytes   `json:"data"`
//		ChainId  hexutil.Uint64  `json:"chain_id"`
//		Payment  *common.Address `json:"payment"`
//		Fee      *hexutil.Big    `json:"fee"`
//	}
//	var enc SignTx
//	enc.Phone = hexutil.Uint64(t.Phone)
//	enc.To = &t.To
//	enc.Value = (*hexutil.Big)(t.Value)
//	enc.GasPrice = hexutil.Uint64(t.GasPrice)
//	enc.GasLimit = hexutil.Uint64(t.GasLimit)
//	enc.Nonce = hexutil.Uint64(t.Nonce)
//	enc.Data = t.Data
//	enc.ChainId = hexutil.Uint64(t.ChainId)
//	enc.Payment = &t.Payment
//	enc.Fee = (*hexutil.Big)(t.Fee)
//	return json.Marshal(&enc)
//}
//
// UnmarshalJSON unmarshals from JSON.
func (h *SignTx) UnmarshalJSON(input []byte) error {
	type SignTx struct {
		Phone    *int64          `json:"userId"`
		To       *common.Address `json:"to"`
		Value    *string         `json:"value"`
		GasPrice *int64          `json:"gasPrice"`
		GasLimit *int64          `json:"gasLimit"`
		Nonce    *int64          `json:"nonce"`
		Data     *hexutil.Bytes  `json:"data"`
		ChainId  *int64          `json:"chainId"`
		Payment  *common.Address `json:"payment"`
	}
	var dec SignTx
	if err := json.Unmarshal(input, &dec); err != nil {
		fmt.Println("UnmarshalJSON ", err)
		return err
	}
	if dec.Phone == nil {
		return errors.New("missing required field 'Phone' for SignTx")
	}
	h.Phone = uint64(*dec.Phone)
	if dec.To != nil {
		h.To = *dec.To
	}
	if dec.Value != nil {
		r, result := new(big.Int).SetString("15151545445646", 10)
		if result {
			return errors.New("missing required field 'Value' can't SetString")
		}
		h.Value = r
	}

	if dec.GasPrice == nil {
		return errors.New("missing required field 'GasPrice' for SignTx")
	}
	h.GasPrice = uint64(*dec.GasPrice)
	if dec.GasLimit == nil {
		return errors.New("missing required field 'GasLimit' for SignTx")
	}
	h.GasLimit = uint64(*dec.GasLimit)
	if dec.Nonce == nil {
		return errors.New("missing required field 'Nonce' for SignTx")
	}
	h.Nonce = uint64(*dec.Nonce)

	if dec.Data == nil {
		return errors.New("missing required field 'Data' for SignTx")
	}
	h.Data = *dec.Data
	if dec.ChainId == nil {
		return errors.New("missing required field 'ChainId' for SignTx")
	}
	h.ChainId = uint64(*dec.ChainId)

	if dec.Payment != nil {
		h.Payment = *dec.Payment
	}
	return nil
}

type AuthResult struct {
	CryptoPub string `json:"crypto_pub"`
}

type AuthQuest struct {
	Hash common.Hash `json:"hash"`
	Sign []byte      `json:"sign"`
}

func (aq AuthQuest) string() string {
	return fmt.Sprintf("[Hash:%s  sign:%s]", aq.Hash.String(), aq.Sign)
}

type AdminQuest struct {
	Root  common.Address `json:"root"`
	Admin common.Address `json:"admin"`
}

type DappQuest struct {
	Name string   `json:"name"`
	IPs  []string `json:"ips"`
	Desc string   `json:"desc"`
}

type DappResult struct {
	ID    common.Hash `json:"dapp_id"`
	Priv  string      `json:"dapp_priv"`
	Index uint64      `json:"index"`
	Note  string      `json:"note"`
}

func (d *DappResult) String() string {
	if d.Note != "" {
		return fmt.Sprintf("[ID:%s Index:%d Note:%s Priv:%s]", d.ID.String(), d.Index, d.Note, d.Priv)
	} else {
		return fmt.Sprintf("[ID:%s Index:%d Priv:%s]", d.ID.String(), d.Index, d.Priv)
	}
}

func (a Account) String() string {
	return fmt.Sprintf("[ID:%d Address:%s Status:%d]\n", a.ID, a.Address.String(), a.Status)
}

func (a AccountResult) String() string {
	return fmt.Sprintf("[Account ID:%s Index:%d Status:%d Ips:%s ]", a.ID.String(), a.Index, a.Status, a.IPs)
}

func (q *QueryResult) String() string {
	var ss string
	ss += "[ ID:" + q.ID.String() + " Index: " + strconv.FormatUint(q.Index, 10) + " Status: " + strconv.FormatUint(q.Status, 10) + " Desc: " + q.Desc
	ss += " Ips: [" + strings.Join(q.IPs, ",") + "]"
	ss += "  priv :" + q.Priv + " ]\n"
	var arrs []string
	for _, v := range q.Ars {
		arrs = append(arrs, v.String())
	}
	ss += " Accounts: [" + strings.Join(arrs, ",") + "]"
	return ss
}

type DeriveQuest struct {
	ID    common.Hash `json:"id"`
	Count uint64      `json:"count"`
	Ips   []string    `json:"ips"`
}

type Account struct {
	ID      uint64         `json:"address_id"`
	Address common.Address `json:"address"` // Ethereum account address derived from the key
	Status  uint64         `json:"address_status"`
}

type UpdateDapppQuest struct {
	ID     common.Hash
	IPs    []string
	Status uint64
	Desc   string
}

type AccountState struct {
	ID        common.Hash    `json:"dapp_id"`
	AddressID common.Address `json:"address_id"`
	IPs       []string       `json:"ips"`
	Status    uint64         `json:"status"`
	Desc      string         `json:"desc"`
}

type DappQuery struct {
	ID        common.Hash    `json:"dapp_id"`
	AddressID common.Address `json:"address_id"`
}

type QueryResult struct {
	ID     common.Hash     `json:"id"`
	Index  uint64          `json:"index"`
	Priv   string          `json:"priv"`
	Status uint64          `json:"status"`
	IPs    []string        `json:"ips"`
	Desc   string          `json:"desc"`
	Ars    []AccountResult `json:"address_list"`
}

type AccountResult struct {
	ID     common.Address `json:"id"`
	Index  uint64         `json:"index"`
	Status uint64         `json:"status"`
	IPs    []string       `json:"ips"`
}
