package types

import (
	"encoding/json"
	"ethereum/keyservice/common"
	"ethereum/keyservice/log"
	"io/ioutil"
	"os"
)

const jsonIndent = "    "

// Config is the config.json file format. It holds a set of node records
// as a JSON object.
type Config struct {
	RpcPort int          `json:"rpcport"`
	RpcAddr string       `json:"rpcaddr"`
	Config  []RootConfig `json:"admins"`
}

type RootConfig struct {
	Root   common.Address   `json:"root"`
	Admins []common.Address `json:"admins"`
}

func LoadNodesJSON(file string) Config {
	var config Config
	if isExist(file) {
		if err := common.LoadJSON(file, &config); err != nil {
			log.Info("loadNodesJSON", "error", err)
		}
	}
	return config
}

func WriteNodesJSON(file string, config Config) {
	for _, v := range LoadNodesJSON(file).Config {
		for _, n := range config.Config {
			if v.Root == n.Root {
				n.Admins = append(n.Admins, v.Admins...)
			}
		}
	}

	nodesJSON, err := json.MarshalIndent(config, "", jsonIndent)
	if err != nil {
		log.Info("writeNodesJSON MarshalIndent", "error", err)
	}
	if file == "-" {
		os.Stdout.Write(nodesJSON)
		return
	}
	if err := ioutil.WriteFile(file, nodesJSON, 0644); err != nil {
		log.Info("writeNodesJSON writeFile", "error", err)
	}
}

func isExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}
