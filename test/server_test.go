package test

import (
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"ethereum/keyservice/crypto"
	"ethereum/keyservice/rpc"
	"ethereum/keyservice/services/truekey/types"
	"fmt"
	"math/big"
	"strings"
	"testing"
)

var (
	adminKey, _ = crypto.HexToECDSA("de492aa324b5f95563dbd7746178fd6328362c9cd676d50a130066876145ab9d")
	admin       = crypto.PubkeyToAddress(adminKey.PublicKey)
	root        = common.HexToAddress("0xc0eb4d5c43e894b42aae58d859cf926afa6a846bd")
)

func TestRegular(t *testing.T) {
	ip := "http://" + "39.100.97.129:8985"
	fmt.Println("ip ", ip)
}

func deriveAccounts(client *rpc.Client, quest types.AdminQuest, id common.Hash) []types.Account {
	var as []types.Account
	var dq types.DeriveQuest
	dq.ID = id
	dq.Count = 6
	dq.Ips = []string{"169.254.1.2", "0.254.1.2"}
	err := client.Call(&as, "truekey_dappDerive", quest, dq)
	if err != nil {
		fmt.Println("account_version Error", err.Error())
		return nil
	}
	for i, account := range as {
		fmt.Println("version", "i", i, "account", account.ID, "address", account.Address.String())
	}
	return as
}

func updateDapp(client *rpc.Client, quest types.AdminQuest, id common.Hash) {
	var v string
	var as types.UpdateDapppQuest
	as.ID = id
	as.Desc = "updateDapp121"
	as.Status = types.Lock
	as.IPs = []string{"169.254.1.2", "1.254.1.2"}
	err := client.Call(&v, "truekey_updateDapp", quest, as)
	if err != nil {
		fmt.Println("updateDapp Error", err.Error())
		return
	}
	fmt.Println("updateDapp version", v)
}

func updateAccount(client *rpc.Client, quest types.AdminQuest, id common.Hash, address common.Address) {
	var v string
	var as types.AccountState
	as.ID = id
	as.AddressID = address
	as.Desc = "updateAccount121"
	as.Status = types.Lock
	as.IPs = []string{"169.254.1.2", "0.254.1.3"}
	err := client.Call(&v, "truekey_updateAccount", quest, as)
	if err != nil {
		fmt.Println("updateAccount Error", err.Error())
		return
	}
	fmt.Println("updateAccount version", v)
}

func dappAddress(client *rpc.Client, quest types.AdminQuest, id common.Hash, address common.Address) {
	var as *types.QueryResult
	var dq types.DappQuery
	dq.ID = id
	dq.AddressID = address
	err := client.Call(&as, "truekey_dappAddress", quest, dq)
	if err != nil {
		fmt.Println("childAddress Error", err.Error())
		return
	}
	for i, account := range as.Ars {
		fmt.Println("version", "i", i, "account", account.ID.String(), "Status", account.Status, "ips", account.IPs)
	}
	fmt.Println("childAddress version", as)
}

func TestDial(t *testing.T) {
	ip := "http://" + "47.92.246.187:8550"
	fmt.Println("ip ", ip)

	client, err := rpc.Dial(ip)
	if err != nil {
		fmt.Println("Dail:", ip, err.Error())
		return
	}
	fmt.Println("admin ", admin.String())
	quest := types.AdminQuest{
		Admin: admin,
		Root:  root,
	}
	dapp := register(client, quest)
	//common.HexToHash("0xab229e93803db84258ee435a6629f5e7522bc51dfbf52d8154f7fb9639d7a724")
	fmt.Println(dapp.String())
}

func register(client *rpc.Client, quest types.AdminQuest) common.Address {
	var v common.Address
	err := client.Call(&v, "truekey_registerAccount", string("18682003824"))
	if err != nil {
		if strings.Contains(err.Error(), types.ErrDappAlready.Error()) {
			return v
		}
		fmt.Println("register Error", err.Error())
		return common.Address{}
	}
	//Payment  common.Address `json:"payment"`
	//Fee      *big.Int       `json:"fee"`
	tx := types.SignTx{
		Phone:    18682003824,
		To:       common.BigToAddress(new(big.Int).SetUint64(100)),
		Value:    big.NewInt(1000),
		GasPrice: 100,
		GasLimit: 1000,
		Nonce:    1,
		ChainId:  100,
	}
	var data hexutil.Bytes

	err = client.Call(&data, "truekey_signHashPlain", tx)
	if err != nil {
		if strings.Contains(err.Error(), types.ErrDappAlready.Error()) {
			return v
		}
		fmt.Println("truekey_signHashPlain Error", err.Error())
		return common.Address{}
	}
	fmt.Println("truekey_register", v.String(), " ", hexutil.Encode(data))
	return v
}
