package main

import (
	"crypto/ecdsa"
	"errors"
	"ethereum/keyservice/accounts/keystore"
	"ethereum/keyservice/common"
	"ethereum/keyservice/console"
	"ethereum/keyservice/crypto"
	"ethereum/keyservice/crypto/ecies"
	"ethereum/keyservice/rlp"
	"ethereum/keyservice/rpc"
	"ethereum/keyservice/services/truekey/types"
	"ethereum/keyservice/services/utils"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var (
	key   string
	store string
	ip    string
	port  int
)

var (
	priKey *ecdsa.PrivateKey
	from   common.Address
)

const (
	datadirDefaultKeyStore = "keystore"
	datadirPrivateKey      = "key"
)

var RegisterCommand = cli.Command{
	Name:   "register",
	Usage:  "Register dapp info",
	Action: utils.MigrateFlags(register),
	Flags:  RegisterFlags,
}

func register(ctx *cli.Context) error {

	loadPrivate(ctx)

	conn, url := dialConn(ctx)
	quest := parseAdminQuestParam(ctx)

	printBaseInfo(conn, quest, url)

	dapp := parseRegisterParam(ctx)
	registerDapp(conn, quest, dapp)
	return nil
}

func parseRegisterParam(ctx *cli.Context) types.DappQuest {
	name := ctx.GlobalString(NameFlag.Name)
	if name == "" {
		printError("Name can't null")
	}
	desc := ctx.GlobalString(DescFlag.Name)
	ips := ctx.GlobalString(IpsFlag.Name)
	dapp := types.DappQuest{
		Name: name,
		Desc: desc,
		IPs:  types.CheckIp(strings.Split(ips, ",")),
	}
	return dapp
}

func parseAdminQuestParam(ctx *cli.Context) types.AdminQuest {
	root := ctx.GlobalString(RootFlag.Name)
	if root == "" {
		printError("Must specify --root")
	}
	if !common.IsHexAddress(root) {
		printError("Please specify correct root address")
	}

	return types.AdminQuest{Root: common.HexToAddress(root),
		Admin: from}
}

//	hashBytes := encryptNode.HashWithoutSign().Bytes()
//	pubKey, err := crypto.SigToPub(hashBytes, encryptNode.Sign)
func registerDapp(client *rpc.Client, quest types.AdminQuest, dapp types.DappQuest) *types.DappResult {
	var v *types.EncryptMessage
	pub := authPub(client, quest)
	if pub == nil {
		fmt.Println("registerDapp auth failed")
		return nil
	}
	encryptQuest, err := signQuest(dapp, pub)
	if err != nil {
		fmt.Println("truekey_registerDapp Error", err.Error())
		return nil
	}
	err = client.Call(&v, "truekey_registerDapp", quest, encryptQuest)
	if err != nil {
		fmt.Println("truekey_registerDapp Error", err.Error())
		return nil
	}
	priKey := ecies.ImportECDSA(priKey)
	decryptMessage, err := priKey.Decrypt(v.DappInfo, nil, nil)
	dappResult := new(types.DappResult)

	if err := rlp.DecodeBytes(decryptMessage, dappResult); err != nil {
		fmt.Println("Failed to decode decrypt message", "err", err)
		return nil
	}
	fmt.Println("truekey registerDapp Success \n", dappResult)
	return dappResult
}

func createKs() {
	ks := keystore.NewKeyStore("./createKs", keystore.StandardScryptN, keystore.StandardScryptP)
	password := "secret"
	account, err := ks.NewAccount(password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(account.Address.Hex()) // 0x20F8D42FB0F667F2E53930fed426f225752453b3
}

func importKs(password string) common.Address {
	file, err := getAllFile(datadirDefaultKeyStore)
	if err != nil {
		log.Fatal(err)
	}
	cks, _ := filepath.Abs(datadirDefaultKeyStore)

	jsonBytes, err := ioutil.ReadFile(filepath.Join(cks, file))
	if err != nil {
		log.Fatal(err)
	}

	//password := "secret"
	key, err := keystore.DecryptKey(jsonBytes, password)
	if err != nil {
		log.Fatal(err)
	}
	priKey = key.PrivateKey
	from = crypto.PubkeyToAddress(priKey.PublicKey)

	fmt.Println("address ", from.Hex())
	return from
}

func getAllFile(path string) (string, error) {
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		printError("path ", err)
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fmt.Printf("[%s]\n", path+"\\"+fi.Name())
			getAllFile(path + fi.Name() + "\\")
			return "", errors.New("path error")
		} else {
			fmt.Println(path, "dir has ", fi.Name(), "file")
			return fi.Name(), nil
		}
	}
	return "", err
}

func printError(error ...interface{}) {
	log.Fatal(error)
}

func loadPrivate(ctx *cli.Context) {
	key = ctx.GlobalString(KeyFlag.Name)
	store = ctx.GlobalString(KeyStoreFlag.Name)
	if key != "" {
		loadPrivateKey(key)
	} else if store != "" {
		loadSigningKey(store)
	} else {
		printError("Must specify --key or --keystore")
	}

	if priKey == nil {
		printError("load privateKey failed")
	}
}

func dialConn(ctx *cli.Context) (*rpc.Client, string) {
	ip = ctx.GlobalString(utils.RPCListenAddrFlag.Name)
	port = ctx.GlobalInt(utils.RPCPortFlag.Name)

	url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	// Create an IPC based RPC connection to a remote node
	// "http://39.100.97.129:8545"
	client, err := rpc.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the TrueKey service: %v", err)
	}
	return client, url
}

func printBaseInfo(conn *rpc.Client, quest types.AdminQuest, url string) {
	fmt.Println("Connect url ", url, " root ", quest.Root.String(), " admin ", from.Hex())
	return
}

func loadPrivateKey(path string) common.Address {
	var err error
	if path == "" {
		file, err := getAllFile(datadirPrivateKey)
		if err != nil {
			printError(" getAllFile file name error", err)
		}
		kab, _ := filepath.Abs(datadirPrivateKey)
		path = filepath.Join(kab, file)
	}
	priKey, err = crypto.LoadECDSA(path)
	if err != nil {
		printError("LoadECDSA error", err)
	}
	from = crypto.PubkeyToAddress(priKey.PublicKey)
	return from
}

// loadSigningKey loads a private key in Ethereum keystore format.
func loadSigningKey(keyfile string) common.Address {
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		printError(fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	password, _ := console.Stdin.PromptPassword("Please enter the password for '" + keyfile + "': ")
	//password := "secret"
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		printError(fmt.Errorf("error decrypting key: %v", err))
	}
	priKey = key.PrivateKey
	from = crypto.PubkeyToAddress(priKey.PublicKey)
	return from
}
