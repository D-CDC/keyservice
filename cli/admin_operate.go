package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"ethereum/keyservice/crypto"
	"ethereum/keyservice/crypto/ecies"
	"ethereum/keyservice/log"
	"ethereum/keyservice/rlp"
	"ethereum/keyservice/rpc"
	"ethereum/keyservice/services/truekey/types"
	"ethereum/keyservice/services/utils"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"strings"
	"time"
)

var DappDeriveCommand = cli.Command{
	Name:   "derive",
	Usage:  "Dapp Derive account",
	Action: utils.MigrateFlags(derive),
	Flags:  DeriveFlags,
}

func derive(ctx *cli.Context) error {
	loadPrivate(ctx)

	conn, url := dialConn(ctx)

	quest := parseAdminQuestParam(ctx)
	printBaseInfo(conn, quest, url)

	dapp := parseDeriveParam(ctx)
	deriveDapp(conn, quest, dapp)

	return nil
}

func parseDeriveParam(ctx *cli.Context) types.DeriveQuest {
	id := ctx.GlobalString(IDFlag.Name)
	if id == "" {
		printError("id can't null")
	}
	count := ctx.GlobalUint64(CountFlag.Name)
	ips := ctx.GlobalString(IpsFlag.Name)
	dapp := types.DeriveQuest{
		ID:    common.HexToHash(id),
		Count: count,
		Ips:   types.CheckIp(strings.Split(ips, ",")),
	}
	return dapp
}

func deriveDapp(client *rpc.Client, quest types.AdminQuest, dapp types.DeriveQuest) []types.Account {
	var v *types.EncryptMessage
	pub := authPub(client, quest)
	if pub == nil {
		fmt.Println("deriveDapp auth failed")
		return nil
	}
	encryptQuest, err := signQuest(dapp, pub)
	if err != nil {
		fmt.Println("truekey_dappDerive Error", err.Error())
		return nil
	}
	err = client.Call(&v, "truekey_dappDerive", quest, encryptQuest)
	if err != nil {
		fmt.Println("truekey_dappDerive Error", err.Error())
		return nil
	}
	priKey := ecies.ImportECDSA(priKey)
	decryptMessage, err := priKey.Decrypt(v.DappInfo, nil, nil)
	dappResult := new([]types.Account)

	if err := rlp.DecodeBytes(decryptMessage, dappResult); err != nil {
		fmt.Println("Failed to decode decrypt message", "err", err)
		return nil
	}
	fmt.Println("truekey dappDerive Success\n", *dappResult)
	return *dappResult
}

func authPub(client *rpc.Client, quest types.AdminQuest) *ecdsa.PublicKey {
	hash := common.HexToHash("hello server")
	sign, err := crypto.Sign(hash.Bytes(), priKey)
	if err != nil {
		log.Error("sign node error", "err", err)
	}
	auth := types.AuthQuest{
		Hash: hash,
		Sign: sign,
	}
	var encryMessage *types.EncryptMessage
	err = client.Call(&encryMessage, "truekey_authPub", quest, auth)
	if err != nil {
		fmt.Println("truekey_authPub Error", err.Error())
		return nil
	}
	priKey := ecies.ImportECDSA(priKey)
	decryptMessage, err := priKey.Decrypt(encryMessage.DappInfo, nil, nil)
	query := new(types.AuthResult)

	if err := rlp.DecodeBytes(decryptMessage, query); err != nil {
		fmt.Println("Failed to decode decrypt message", "err", err)
		return nil
	}
	pubData, err := hexutil.Decode(query.CryptoPub)
	if err != nil {
		fmt.Println("truekey_authPub decode Error", err.Error())
		return nil
	}
	pub, err := crypto.UnmarshalPubkey(pubData)
	if err != nil {
		fmt.Println("truekey_authPub unmarshal Error", err.Error())
		return nil
	}
	return pub
}

func signQuest(val interface{}, pub *ecdsa.PublicKey) (*types.EncryptMessage, error) {
	resultByte, err := rlp.EncodeToBytes(val)
	if err != nil {
		log.Error("EncodeToBytes error: ", "err", err)
		return nil, err
	}
	cryMessage := &types.EncryptMessage{
		CreatedAt: hexutil.Uint64(time.Now().Unix()),
	}
	encryptMessageInfo, err := ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(pub), resultByte, nil, nil)
	if err != nil {
		log.Error("publickey encrypt result error ", "publickey", common.Bytes2Hex(crypto.FromECDSAPub(pub)), "err", err)
		return nil, err
	}
	cryMessage.DappInfo = encryptMessageInfo
	hash := cryMessage.HashWithoutSign().Bytes()
	cryMessage.Sign, err = crypto.Sign(hash, priKey)
	return cryMessage, nil
}

var UpdateDappCommand = cli.Command{
	Name:   "updatedapp",
	Usage:  "Update dapp config",
	Action: utils.MigrateFlags(UpdateDapp),
	Flags:  UpdateDappFlags,
}

func UpdateDapp(ctx *cli.Context) error {
	loadPrivate(ctx)

	conn, url := dialConn(ctx)

	quest := parseAdminQuestParam(ctx)
	printBaseInfo(conn, quest, url)

	dapp := parseUpdateDappParam(ctx)
	updateDapppCall(conn, quest, dapp)

	return nil
}

func parseUpdateDappParam(ctx *cli.Context) types.UpdateDapppQuest {
	id := ctx.GlobalString(IDFlag.Name)
	if id == "" {
		printError("id can't null")
	}
	status := ctx.GlobalUint64(StatusFlag.Name)
	ips := ctx.GlobalString(IpsFlag.Name)
	desc := ctx.GlobalString(DescFlag.Name)
	dapp := types.UpdateDapppQuest{
		ID:     common.HexToHash(id),
		Status: status,
		IPs:    types.CheckIp(strings.Split(ips, ",")),
		Desc:   desc,
	}
	return dapp
}

func updateDapppCall(client *rpc.Client, quest types.AdminQuest, dapp types.UpdateDapppQuest) {
	var v string
	pub := authPub(client, quest)
	if pub == nil {
		fmt.Println("updateDappp auth failed")
		return
	}
	encryptQuest, err := signQuest(dapp, pub)
	if err != nil {
		fmt.Println("truekey_updateDapp Error", err.Error())
		return
	}
	err = client.Call(&v, "truekey_updateDapp", quest, encryptQuest)
	if err != nil {
		fmt.Println("truekey_updateDapp Error", err.Error())
		return
	}
	fmt.Println("truekey updateDappp Success\n ")
	return
}

var UpdateAccountCommand = cli.Command{
	Name:   "updateaccount",
	Usage:  "Update account",
	Action: utils.MigrateFlags(updateAccount),
	Flags:  UpdateAccountFlags,
}

func updateAccount(ctx *cli.Context) error {
	loadPrivate(ctx)

	conn, url := dialConn(ctx)

	quest := parseAdminQuestParam(ctx)
	printBaseInfo(conn, quest, url)

	dapp := parseUpdateAccountParam(ctx)
	updateAccountCall(conn, quest, dapp)

	return nil
}

func parseUpdateAccountParam(ctx *cli.Context) types.AccountState {
	id := ctx.GlobalString(IDFlag.Name)
	if id == "" {
		printError("id can't null")
	}
	status := ctx.GlobalUint64(StatusFlag.Name)
	ips := ctx.GlobalString(IpsFlag.Name)
	desc := ctx.GlobalString(DescFlag.Name)
	address := ctx.GlobalString(AddressFlag.Name)
	if !common.IsHexAddress(address) {
		printError("Must input correct address")
	}
	dapp := types.AccountState{
		ID:        common.HexToHash(id),
		AddressID: common.HexToAddress(address),
		Status:    status,
		IPs:       types.CheckIp(strings.Split(ips, ",")),
		Desc:      desc,
	}
	return dapp
}

func updateAccountCall(client *rpc.Client, quest types.AdminQuest, dapp types.AccountState) {
	var v string
	pub := authPub(client, quest)
	if pub == nil {
		fmt.Println("updateAccount auth failed")
		return
	}
	encryptQuest, err := signQuest(dapp, pub)
	if err != nil {
		fmt.Println("truekey_updateAccount Error", err.Error())
		return
	}
	err = client.Call(&v, "truekey_updateAccount", quest, encryptQuest)
	if err != nil {
		fmt.Println("truekey_updateAccount Error", err.Error())
		return
	}
	fmt.Println("truekey updateAccount Success\n ")
	return
}

var DappAddressCommand = cli.Command{
	Name:   "dappaddress",
	Usage:  "Dapp list a account or all account",
	Action: utils.MigrateFlags(dappAddress),
	Flags:  DappAddressFlags,
}

func dappAddress(ctx *cli.Context) error {
	loadPrivate(ctx)

	conn, url := dialConn(ctx)

	quest := parseAdminQuestParam(ctx)
	printBaseInfo(conn, quest, url)

	dapp := parseDappAddressParam(ctx)
	dappAddressCall(conn, quest, dapp)

	return nil
}

func parseDappAddressParam(ctx *cli.Context) types.DappQuery {
	var id common.Hash
	var address common.Address

	if ctx.GlobalIsSet(IDFlag.Name) {
		id = common.HexToHash(ctx.GlobalString(IDFlag.Name))
	} else {
		id = common.Hash{}
	}

	if ctx.GlobalIsSet(AddressFlag.Name) {
		addressS := ctx.GlobalString(AddressFlag.Name)
		if !common.IsHexAddress(addressS) {
			printError("Must input correct address")
		}
		address = common.HexToAddress(addressS)
	} else {
		address = common.Address{}
	}

	dapp := types.DappQuery{
		ID:        id,
		AddressID: address,
	}
	return dapp
}

func dappAddressCall(client *rpc.Client, quest types.AdminQuest, dapp types.DappQuery) []*types.QueryResult {
	var v *types.EncryptMessage
	pub := authPub(client, quest)
	if pub == nil {
		fmt.Println("dappAddress auth failed")
		return nil
	}
	encryptQuest, err := signQuest(dapp, pub)
	if err != nil {
		fmt.Println("truekey_dappAddress Error", err.Error())
		return nil
	}
	err = client.Call(&v, "truekey_dappAddress", quest, encryptQuest)
	if err != nil {
		fmt.Println("truekey_dappAddress Error", err.Error())
		return nil
	}
	priKey := ecies.ImportECDSA(priKey)
	decryptMessage, err := priKey.Decrypt(v.DappInfo, nil, nil)
	var dappResult []*types.QueryResult

	if err := rlp.Decode(bytes.NewReader(decryptMessage), &dappResult); err != nil {
		fmt.Println("Failed to decode decrypt message", "err", err)
		return nil
	}
	fmt.Println("truekey dappAddress Success\n ")
	for _, v := range dappResult {
		fmt.Println(v)
		fmt.Println()
	}
	return dappResult
}
