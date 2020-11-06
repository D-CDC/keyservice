package types

import (
	"ethereum/keyservice/common"
	"ethereum/keyservice/rlp"
	"fmt"
	"testing"
)

func TestAdminWallet(t *testing.T) {

	hash := common.HexToAddress("2121")
	wallet, _ := NewAdminWallet(hash)
	bzs, err := rlp.EncodeToBytes(wallet)
	if err != nil {
		fmt.Println(err.Error())
	}

	var tmp AdminWallet
	err = rlp.DecodeBytes(bzs, &tmp)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("account %s \n", tmp.Address)
}

func TestDappInfo(t *testing.T) {
	hash := common.HexToAddress("2121")
	quest := DappQuest{Name: "test", IPs: []string{"169.254.1.2", "0.254.1.2"},
		Desc: "desc"}
	wallet, _ := NewDappIdentify(quest, hash)
	bzs, err := rlp.EncodeToBytes(wallet)
	if err != nil {
		fmt.Println(err.Error())
	}

	var tmp DappIdentify
	err = rlp.DecodeBytes(bzs, &tmp)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("account %s \n", tmp.Desc)
}

func TestMap(t *testing.T) {
	map1 := make(map[int]int)
	map1[1] = 1
	map1[2] = 2
	map1[3] = 3
	map2 := make(map[int]int)
	map2[1] = 1
	map2[2] = 2
	map1 = map2
	fmt.Println(map1)
}
