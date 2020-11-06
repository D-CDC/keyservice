package test

import (
	"ethereum/keyservice/common"
	"ethereum/keyservice/rpc"
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	diaAcount()
}

func diaAcount() {
	ip := "http://" + "47.241.32.76:8550"
	client, err := rpc.Dial(ip)
	if err != nil {
		fmt.Println("Dail:", ip, err.Error())
		return
	}
	version(client)
	listAccount(client, common.HexToHash("12"))
}

func listAccount(client *rpc.Client, key common.Hash) {
	var v []common.Address
	err := client.Call(&v, "account_authDapp", key)
	if err != nil {
		fmt.Println("account_authDapp Error", err.Error())
		return
	}
	fmt.Println("account_authDapp ", v)
}

func version(client *rpc.Client) {
	var v string
	err := client.Call(&v, "account_version")
	if err != nil {
		fmt.Println("account_register Error", err.Error())
		return
	}
	fmt.Println("version ", v)
}

func listWallet(client *rpc.Client) []common.Hash {
	var v []common.Hash
	err := client.Call(&v, "account_listWallet")
	if err != nil {
		fmt.Println("account_listWallet Error", err.Error())
		return nil
	}
	fmt.Println("listWallet ", v)
	return v
}
