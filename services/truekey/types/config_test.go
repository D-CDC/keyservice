package types

import (
	"ethereum/keyservice/common"
	"ethereum/keyservice/crypto"
	"fmt"
	"testing"
)

func TestWriteNodes(t *testing.T) {
	priKey, _ := crypto.HexToECDSA("0260c952edc49037129d8cabbe4603d15185d83aa718291279937fb6db0fa7a2")
	root1 := crypto.PubkeyToAddress(priKey.PublicKey)

	rootkey2, _ := crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	root2 := crypto.PubkeyToAddress(rootkey2.PublicKey)

	var addrs []common.Address
	for i := 0; i < 4; i++ {
		key, _ := crypto.GenerateKey()
		addrs = append(addrs, crypto.PubkeyToAddress(key.PublicKey))
	}

	configAdmins := LoadNodesJSON("config.json")
	fmt.Println("configAdmins", configAdmins, " configAdmins ", configAdmins.RpcPort, configAdmins.RpcAddr, configAdmins.Config)

	WriteNodesJSON("config.json", Config{
		8985,
		"127.0.0.1",
		[]RootConfig{
			{
				Root: root1,
				Admins: []common.Address{
					addrs[0],
					addrs[1],
				},
			}, {Root: root2,
				Admins: []common.Address{
					addrs[2],
					addrs[3],
				},
			},
		},
	})
}
